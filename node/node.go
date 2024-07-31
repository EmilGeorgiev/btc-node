package node

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

// Node is a central part in the program that hold reference to all Peer and manage communication with them.
type Node struct {
	newServerPeer          func(p2p.Peer, chan PeerErr) PeerConnectionManager
	network                string
	userAgent              string
	peerChain              *sync.Map
	errors                 chan PeerErr
	syncCompleted          chan struct{}
	peerAddrs              []common.Addr
	handshakeManager       HandshakeManager
	getNextPeerConnMngWait time.Duration
	reconnectWait          time.Duration

	stop               chan struct{}
	wg                 *sync.WaitGroup
	notifySyncForError chan PeerErr
}

// New initialize and return a new Node.
func New(network, userAgent string, newServerPeer func(p2p.Peer, chan PeerErr) PeerConnectionManager,
	peerAddr []common.Addr, err chan PeerErr, sf chan struct{}, hm HandshakeManager, w time.Duration, recWait time.Duration) (*Node, error) {
	_, ok := p2p.Networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	return &Node{
		newServerPeer:          newServerPeer,
		network:                network,
		userAgent:              userAgent,
		peerAddrs:              peerAddr,
		errors:                 err,
		peerChain:              &sync.Map{},
		syncCompleted:          sf,
		handshakeManager:       hm,
		getNextPeerConnMngWait: w,
		stop:                   make(chan struct{}, 1000),
		notifySyncForError:     make(chan PeerErr, 1000),
		wg:                     &sync.WaitGroup{},
		reconnectWait:          recWait,
	}, nil
}

// Start this method initialize a connections to all the peers provided in a list.
// Then it scan network in these peers and choose the one with the greatest cumulative PoW.
func (n *Node) Start() {
	log.Println("Start Node.")
	if len(n.peerAddrs) == 0 {
		log.Println("At least one peer address should be provided. Stop the node")
		return
	}
	n.wg.Add(1)
	go n.listenForPeerErrors()
	for _, peerAddr := range n.peerAddrs {
		err := n.connectToPeer(peerAddr)
		if err == nil {
			continue
		}
		n.peerChain.Store(peerAddr, PeerChain{})
		n.errors <- PeerErr{Peer: p2p.Peer{Address: peerAddr.String()}, Err: err}
	}
}

// getChainOverview get overview ( number of blocks, whether the blocks are valid and so others) of the chain of the current peer.
func (n *Node) getChainOverview(pch PeerChain) {
	defer n.wg.Done()
	log.Println("Get Overview of the peer's chain: ", pch.peer.GetPeerAddr())
	overviewCh, err := pch.peer.GetChainOverview()
	if err != nil {
		log.Println("Receive an error from GetChainOverview:", err)
		return
	}
	for {
		select {
		case <-n.stop:
			return
		case view, ok := <-overviewCh:
			if !ok {
				return
			}
			log.Println("Overview for the peer is done: ", pch.peer.GetPeerAddr())
			pch.view = &view
			log.Printf("Overview: %#v\n", pch.view)
			n.peerChain.Store(pch.peer.GetPeerAddr(), pch)
			n.selectBestPeerChainForSync()
			return
		}
	}
}

type PeerChain struct {
	peer PeerConnectionManager
	view *common.ChainOverview
}

func (n *Node) selectBestPeerChainForSync() {
	chainsOverviewsCompleted := true
	var bestChain PeerChain
	n.peerChain.Range(func(key, value any) bool {
		pch := value.(PeerChain)
		if pch.view == nil {
			chainsOverviewsCompleted = false
			return false
		}

		if bestChain.view == nil {
			bestChain = pch
			return true
		}

		if bestChain.view.CumulativeWork.Cmp(pch.view.CumulativeWork) == -1 {
			bestChain = pch
		}

		return true
	})

	if !chainsOverviewsCompleted {
		log.Println("overview is nit completed")
		return
	}

	log.Printf("THE BEST CHAIN is from PEER: %#v\n", bestChain.peer.GetPeerAddr())
	bestChain.peer.Sync()
}

func (n *Node) Stop() {
	log.Println("Stop Node.")
	close(n.stop)
	n.wg.Wait()
	n.peerChain.Range(func(key, value any) bool {
		pch := value.(PeerChain)
		pch.peer.Stop()
		return true
	})
	log.Println("all goroutines are stopped")
}

func (n *Node) reconnectToPeer(addr common.Addr) {
	defer n.wg.Done()
	seconds := n.reconnectWait
	timer := time.NewTimer(seconds)
	for {
		select {
		case <-n.stop:
			log.Println("Stop reconnect logic")
			return
		case <-timer.C:
			log.Println("Try to Reconnect to peer: ", addr.String())
			if err := n.connectToPeer(addr); err == nil {
				log.Println("Stop reconnect logic becasue connect success")
				return
			}

			if n.reconnectWait < 7200 {
				seconds = seconds * 2
			}
			log.Printf("reconnect to peer %s fail. The node will try again after %d seconds", addr.String(), seconds)
			timer = time.NewTimer(seconds)
		}
	}
}

func (n *Node) connectToPeer(addr common.Addr) error {
	handshake, err := n.handshakeManager.CreateOutgoingHandshake(addr, n.network, n.userAgent)
	if err != nil {
		return err
	}

	pcm := n.newServerPeer(handshake.Peer, n.errors)
	pch := PeerChain{peer: pcm}
	n.peerChain.Store(addr.String(), pch)
	pcm.Start()

	n.wg.Add(1)
	go n.getChainOverview(pch)

	return nil
}

func (n *Node) listenForPeerErrors() {
	defer n.wg.Done()
	log.Println("Start Node's listener for peer error.")
	for {
		select {
		case <-n.stop:
			log.Println("Stop Node's listener for peer error.")
			return
		case peerErr := <-n.errors:
			log.Println("ERROR Listener: receive err: ", peerErr.Err)
			addr := peerErr.Peer.Address
			if _, ok := n.peerChain.Load(addr); !ok {
				log.Println("doesn't exists peer with IP:", addr)
				continue
			}

			n.peerChain.Delete(addr)
			address := common.AddrFromString(addr)

			n.wg.Add(1)
			go n.reconnectToPeer(address)
		}
	}
}

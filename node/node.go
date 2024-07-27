package node

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

//var genesisBlockHash = [32]byte{
//	0x00, 0x00, 0x00, 0x00, 0x00, 0x19, 0xd6, 0x68,
//	0x9c, 0x08, 0x5a, 0xe1, 0x65, 0x83, 0x1e, 0x93,
//	0x4f, 0xf7, 0x63, 0xae, 0x46, 0xa2, 0xa6, 0xc1,
//	0x72, 0xb3, 0xf1, 0xb6, 0x0a, 0x8c, 0xe2, 0x6f,
//}

type Node struct {
	newPeerConnectionMng   func(p2p.Peer, chan PeerErr) PeerConnectionManager
	network                string
	userAgent              string
	serverPeer             *sync.Map
	errors                 chan PeerErr
	syncCompleted          chan struct{}
	peerAddrs              []common.Addr
	handshakeManager       HandshakeManager
	getNextPeerConnMngWait time.Duration

	stop               chan struct{}
	doneErrorListener  chan struct{}
	doneSync           chan struct{}
	notifySyncForError chan PeerErr
}

// New returns a new Node.
func New(network, userAgent string, newServerPeer func(p2p.Peer, chan PeerErr) PeerConnectionManager,
	peerAddr []common.Addr, err chan PeerErr, sf chan struct{}, hm HandshakeManager, w time.Duration) (*Node, error) {
	_, ok := p2p.Networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	return &Node{
		newPeerConnectionMng:   newServerPeer,
		network:                network,
		userAgent:              userAgent,
		peerAddrs:              peerAddr,
		errors:                 err,
		serverPeer:             &sync.Map{},
		syncCompleted:          sf,
		handshakeManager:       hm,
		getNextPeerConnMngWait: w,
		stop:                   make(chan struct{}),
		doneErrorListener:      make(chan struct{}, 10),
		doneSync:               make(chan struct{}, 10),
		notifySyncForError:     make(chan PeerErr, 100),
	}, nil
}

func (n *Node) Start() {
	if len(n.peerAddrs) == 0 {
		log.Println("At least one peer address should be provided. Stop the node")
		return
	}
	go n.listenForPeerErrors()
	for _, peerAddr := range n.peerAddrs {
		if err := n.connectToPeer(peerAddr); err != nil {
			n.errors <- PeerErr{
				Peer: p2p.Peer{Address: peerAddr.String()},
				Err:  err,
			}
		}
	}
	go n.syncPeers()
}

func (n *Node) Stop() {
	close(n.stop)
	<-n.doneErrorListener // listen for errors
	<-n.doneSync          // listen for sync

	n.serverPeer.Range(func(key, value any) bool {
		pnm := value.(PeerConnectionManager)
		pnm.Stop()
		return true
	})

	log.Println("all goroutines are stopped")
}

func (n *Node) reconnectPeer(addr common.Addr) {
	seconds := 1
	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	for {
		select {
		case <-n.stop:
			return
		case <-timer.C:
			if err := n.connectToPeer(addr); err == nil {
				return
			}

			if seconds < 7200 {
				seconds = seconds * 2
			}
			log.Printf("reconnect to peer %s faile. The node will try again after %d seconds", addr.String(), seconds)
			timer = time.NewTimer(time.Duration(seconds) * time.Second)
		}
	}
}

func (n *Node) connectToPeer(addr common.Addr) error {
	handshake, err := n.handshakeManager.CreateOutgoingHandshake(addr, n.network, n.userAgent)
	if err != nil {
		return err
	}

	pcm := n.newPeerConnectionMng(handshake.Peer, n.errors)
	n.serverPeer.Store(addr.String(), pcm)
	pcm.Start()
	return nil
}

func (n *Node) listenForPeerErrors() {
	for {
		select {
		case <-n.stop:
			log.Println("stop goroutine that listen for peer errors.")
			n.doneErrorListener <- struct{}{}
			return
		case peerErr := <-n.errors:
			addr := peerErr.Peer.Address
			sp, ok := n.serverPeer.Load(addr)
			if !ok {
				log.Println("doesn't exists peer with IP:", addr)
				continue
			}
			pcm := sp.(PeerConnectionManager)
			pcm.Stop()
			n.serverPeer.Delete(addr)
			n.notifySyncForError <- peerErr

			p := strings.Split(addr, ":")
			if len(p) != 2 {
				log.Printf("Invalid peer IP address: %s", addr)
				continue
			}
			port, err := strconv.Atoi(p[1])
			if err != nil {
				log.Printf("Invalid port number if IP address: %s", addr)
				continue
			}
			address := common.Addr{
				IP:   p[0],
				Port: int64(port),
			}

			go n.reconnectPeer(address)
		}
	}
}

func (n *Node) getNextPeerConnMngForSync(currentIndex int) (PeerConnectionManager, int) {
	tick := time.Tick(n.getNextPeerConnMngWait)
	for {
		select {
		case <-n.stop:
			return nil, 0
		case <-tick:
			currentIndex++
			if currentIndex >= len(n.peerAddrs) {
				currentIndex = 0 // reset the index and start from beginning of the list
			}

			// check for available PeerConnectors to the end of the slice.
			for i := currentIndex; i < len(n.peerAddrs); i++ {
				addr := n.peerAddrs[i]
				v, ok := n.serverPeer.Load(addr.String())
				if ok {
					newPeerConnMng := v.(PeerConnectionManager)
					return newPeerConnMng, i
				}
			}

			// check for available PeerConnectors to from the beginning of the slice .
			for i := 0; i < len(n.peerAddrs); i++ {
				addr := n.peerAddrs[i]
				v, ok := n.serverPeer.Load(addr.String())
				if ok {
					newPeerConnMng := v.(PeerConnectionManager)
					return newPeerConnMng, i
				}
			}
		}
	}
}

func (n *Node) syncPeers() {
	var currentPeerConnMng PeerConnectionManager
	currentIndex := -1
Loop:
	for {
		currentPeerConnMng, currentIndex = n.getNextPeerConnMngForSync(currentIndex)
		if currentPeerConnMng == nil {
			// this means that the method Stop is alled
			n.doneSync <- struct{}{}
			return
		}
		currentPeerConnMng.Sync()
		for {
			select {
			case <-n.stop:
				log.Println("stop goroutine that manage Sync peers.")
				currentPeerConnMng.StopSync()
				n.doneSync <- struct{}{}
				return
			case peerErr := <-n.notifySyncForError:
				if currentPeerConnMng.GetPeerAddr() == peerErr.Peer.Address {
					currentPeerConnMng.StopSync()
					// if current peer to ehih w sync has error we continue to the next one
					continue Loop
				}
			case <-n.syncCompleted:
				currentPeerConnMng.StopSync()
				continue Loop
			}
		}
	}
}

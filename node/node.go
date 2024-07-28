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

type Node struct {
	newServerPeer          func(p2p.Peer, chan PeerErr) StartStop
	network                string
	userAgent              string
	serverPeer             *sync.Map
	errors                 chan PeerErr
	syncCompleted          chan struct{}
	peerAddrs              []common.Addr
	handshakeManager       HandshakeManager
	getNextPeerConnMngWait time.Duration

	stop               chan struct{}
	wg                 *sync.WaitGroup
	notifySyncForError chan PeerErr
}

// New returns a new Node.
func New(network, userAgent string, newServerPeer func(p2p.Peer, chan PeerErr) StartStop,
	peerAddr []common.Addr, err chan PeerErr, sf chan struct{}, hm HandshakeManager, w time.Duration) (*Node, error) {
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
		serverPeer:             &sync.Map{},
		syncCompleted:          sf,
		handshakeManager:       hm,
		getNextPeerConnMngWait: w,
		stop:                   make(chan struct{}, 1000),
		notifySyncForError:     make(chan PeerErr, 1000),
		wg:                     &sync.WaitGroup{},
	}, nil
}

func (n *Node) Start() {
	log.Println("Start Node.")
	if len(n.peerAddrs) == 0 {
		log.Println("At least one peer address should be provided. Stop the node")
		return
	}
	n.wg.Add(1)
	go n.listenForPeerErrors()
	for _, peerAddr := range n.peerAddrs {
		if err := n.connectToPeer(peerAddr); err != nil {
			n.errors <- PeerErr{
				Peer: p2p.Peer{Address: peerAddr.String()},
				Err:  err,
			}
		}
	}
}

func (n *Node) Stop() {
	log.Println("Stop Node.")
	close(n.stop)
	n.wg.Wait()
	log.Println("Stop Node. 111111111")
	n.serverPeer.Range(func(key, value any) bool {
		log.Println("Stop Node. 222222")
		pnm := value.(StartStop)
		pnm.Stop()
		log.Println("Stop Node.3333333")
		return true
	})
	log.Println("Stop Node.444444")
	log.Println("all goroutines are stopped")
}

func (n *Node) reconnectToPeer(addr common.Addr) {
	defer n.wg.Done()
	log.Println("START Reconnect to peer", addr.String())
	seconds := 4
	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	for {
		select {
		case <-n.stop:
			log.Println("Stop reconnect logic")
			return
		case <-timer.C:
			if err := n.connectToPeer(addr); err == nil {
				log.Println("Stop reconnect logic becasue connect success")
				return
			}

			if seconds < 7200 {
				seconds = seconds * 2
			}
			log.Printf("reconnect to peer %s fail. The node will try again after %d seconds", addr.String(), seconds)
			timer = time.NewTimer(time.Duration(seconds) * time.Second)
		}
	}
}

func (n *Node) connectToPeer(addr common.Addr) error {
	handshake, err := n.handshakeManager.CreateOutgoingHandshake(addr, n.network, n.userAgent)
	if err != nil {
		return err
	}

	pcm := n.newServerPeer(handshake.Peer, n.errors)
	n.serverPeer.Store(addr.String(), pcm)
	pcm.Start()
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
			if _, ok := n.serverPeer.Load(addr); !ok {
				log.Println("doesn't exists peer with IP:", addr)
				continue
			}

			n.serverPeer.Delete(addr)
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

			n.wg.Add(1)
			go n.reconnectToPeer(address)
		}
	}
}

//func (n *Node) getNextPeerConnMngForSync() PeerConnectionManager {
//seconds := 1
//tick := time.NewTimer(seconds*time.Second)
//var pcm PeerConnectionManager
//
//for {
//	select {
//	case <-n.stop:
//		return nil
//	case <-tick:
//		pcm = nil
//
//	}
//}
//}
//
//func (n *Node) syncPeers() {
//	log.Println("Start Node's sync peer logic.")
//	var pcm PeerConnectionManager
//	for {
//		n.serverPeer.Range(func(key, value any) bool {
//			pcm = value.(PeerConnectionManager)
//			return false
//		})
//		if pcm == nil {
//			continue
//		}
//		pcm.Sync()
//		select {
//		case <-n.stop:
//			log.Println("Stop Node's sync peer logic.")
//			if pcm != nil {
//				pcm.StopSync()
//			}
//			n.doneSync <- struct{}{}
//			return
//		case peerErr := <-n.notifySyncForError:
//			if peerErr.Peer.Address == pcm.GetPeerAddr() {
//				pcm.StopSync()
//			}
//		case <-n.syncCompleted:
//
//		}
//	}
//	//	log.Println("Sync with peersssssss")
//	//	var currentPeerConnMng PeerConnectionManager
//	//	currentIndex := -1
//	//Loop:
//	//	for {
//	//		log.Println("check for available ServerPeers")
//	//		currentPeerConnMng, currentIndex = n.getNextPeerConnMngForSync(currentIndex)
//	//		if currentPeerConnMng == nil {
//	//			// this means that the method Stop is alled
//	//			n.doneSync <- struct{}{}
//	//			return
//	//		}
//	//		log.Println("NODE SYNC PEER: start SYNC. call peerconnmanager")
//	//		currentPeerConnMng.Sync()
//	//		for {
//	//			select {
//	//			case <-n.stop:
//	//				log.Println("stop goroutine that manage Sync peers.")
//	//				currentPeerConnMng.StopSync()
//	//				n.doneSync <- struct{}{}
//	//				return
//	//			case peerErr := <-n.notifySyncForError:
//	//				if currentPeerConnMng.GetPeerAddr() == peerErr.Peer.Address {
//	//					log.Println("stop syncing with peer")
//	//					currentPeerConnMng.StopSync()
//	//					// if current peer to ehih w sync has error we continue to the next one
//	//					continue Loop
//	//				}
//	//			case <-n.syncCompleted:
//	//				fmt.Println("sync completed")
//	//				currentPeerConnMng.StopSync()
//	//				continue Loop
//	//			}
//	//		}
//	//	}
//}

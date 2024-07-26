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
	newServerPeer func(p2p.Peer, chan<- PeerErr) *ServerPeer
	network       string
	userAgent     string
	serverPeer    sync.Map
	errors        chan PeerErr
	syncCompleted chan string
	peerAddrs     []common.Addr

	stop chan struct{}
	done chan struct{}
}

// New returns a new Node.
func New(network, userAgent string, newServerPeer func(p2p.Peer, chan<- PeerErr) *ServerPeer,
	peerAddr []common.Addr, err chan PeerErr, sf chan string) (*Node, error) {
	_, ok := p2p.Networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	return &Node{
		newServerPeer: newServerPeer,
		network:       network,
		userAgent:     userAgent,
		peerAddrs:     peerAddr,
		errors:        err,
		serverPeer:    sync.Map{},
		syncCompleted: sf,
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}, nil
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
	handshake, err := p2p.CreateOutgoingHandshake(addr, n.network, n.userAgent)
	if err != nil {
		return err
	}

	sp := n.newServerPeer(handshake.Peer, n.errors)
	n.serverPeer.Store(sp.peer.Address, sp)
	sp.Start()
	return nil
}

func (n *Node) Start() {
	go n.managePeers()
	for _, peerAddr := range n.peerAddrs {
		if err := n.connectToPeer(peerAddr); err != nil {
			n.errors <- PeerErr{
				peer: p2p.Peer{Address: peerAddr.String()},
				err:  err,
			}
		}
	}
	go n.syncPeers()
}

func (n *Node) Stop() {
	close(n.stop)
	<-n.done
	log.Println("all goroutines are stopped")
}

func (n *Node) managePeers() {
	for {
		select {
		case <-n.stop:
			log.Println("stop goroutine that manage peers in Node.")
			n.done <- struct{}{}
			return
		case peerErr := <-n.errors:
			addr := peerErr.peer.Address
			sp, ok := n.serverPeer.Load(addr)
			if !ok {
				log.Println("doesn't exists peer with IP:", addr)
				continue
			}
			sp.(*ServerPeer).Stop()
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

			go n.reconnectPeer(address)
		}
	}
}

func (n *Node) syncPeers() {
	var currentPeerSync *ServerPeer
	n.serverPeer.Range(func(key, value any) bool {
		sp := value.(*ServerPeer)
		sp.Sync()
		currentPeerSync = sp
		return false
	})
	for {
		select {
		case <-n.stop:
			log.Println("stop goroutine that manage peers in Node.")
			n.done <- struct{}{}
			return
		case <-n.syncCompleted:
			if currentPeerSync == nil {
				continue
			}
			currentPeerSync.StopSync()
			n.serverPeer.Range(func(key, value any) bool {
				if key == currentPeerSync.peer.Address {
					return true
				}
				currentPeerSync = value.(*ServerPeer)
				return false
			})
			currentPeerSync.Sync()
		}
	}
}

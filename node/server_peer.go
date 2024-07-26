package node

import (
	"errors"
	"fmt"
	errors2 "github.com/EmilGeorgiev/btc-node/errors"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"log"
	"net"
	"sync/atomic"
)

type NetworkMessageHandler interface {
	ReadMessage(conn net.Conn) (interface{}, error)
	WriteMessage(msg *p2p.Message, conn net.Conn) error
}

type ServerPeer struct {
	msgHandlersManager    StartStop
	peerSync              StartStop
	networkMessageHandler NetworkMessageHandler
	peer                  p2p.Peer
	outgoingMsgs          chan *p2p.Message
	errors                chan<- PeerErr
	isStarted             atomic.Bool
	network               string

	msgHeaders chan<- *p2p.MsgHeaders
	msgBlocks  chan<- *p2p.MsgBlock

	stop chan struct{}
	done chan struct{}
}

func NewServerPeer(network string, mhm StartStop, ps StartStop, nmh NetworkMessageHandler, p p2p.Peer,
	out chan *p2p.Message, e chan<- PeerErr, h chan<- *p2p.MsgHeaders, b chan<- *p2p.MsgBlock) *ServerPeer {
	return &ServerPeer{
		network:               network,
		msgHandlersManager:    mhm,
		peerSync:              ps,
		networkMessageHandler: nmh,
		peer:                  p,
		outgoingMsgs:          out,
		errors:                e,
		msgHeaders:            h,
		msgBlocks:             b,
		stop:                  make(chan struct{}),
		done:                  make(chan struct{}),
	}
}

func (sp *ServerPeer) Start() {
	if sp.isStarted.Load() {
		return
	}
	sp.isStarted.Store(true)
	sp.msgHandlersManager.Start()
	go sp.handleIncomingMsgs()
	go sp.handOutgoingMsgs()
}

func (sp *ServerPeer) Sync() {
	sp.peerSync.Start()
}

func (sp *ServerPeer) StopSync() {
	sp.peerSync.Stop()
}

func (sp *ServerPeer) Stop() {
	sp.peerSync.Stop()
	sp.msgHandlersManager.Stop()
	close(sp.stop)
	<-sp.done // waiting for the goroutine that read from the conn to stop
	<-sp.done // waiting for the goroutine that write to the conn to stop
}

type PeerErr struct {
	peer p2p.Peer
	err  error
}

func (sp *ServerPeer) handleIncomingMsgs() {
	conn := sp.peer.Connection
	addr := sp.peer.Address
	for {
		select {
		case <-sp.stop:
			log.Println("Stop goroutine that handle messages from peer: ", addr)
			sp.done <- struct{}{}
			return
		default:
			msg, err := sp.networkMessageHandler.ReadMessage(conn)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					log.Println("timeout read")
					continue
				}
				sp.errors <- PeerErr{
					peer: sp.peer,
					err:  errors2.NewE(fmt.Sprintf("receive an error while reading from peer: %s.", addr), err, true),
				}
				continue
			}
			sp.handleMessage(msg)
		}
	}
}

func (sp *ServerPeer) handOutgoingMsgs() {
	conn := sp.peer.Connection
	addr := sp.peer.Address
	for {
		select {
		case <-sp.stop:
			log.Println("Stop goroutine that handle outgoing msg for peers: ", addr)
			sp.done <- struct{}{}
			return
		case msg := <-sp.outgoingMsgs:
			err := sp.networkMessageHandler.WriteMessage(msg, conn)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					log.Println("timeout read")
					continue
				}
				sp.errors <- PeerErr{
					peer: sp.peer,
					err:  errors2.NewE(fmt.Sprintf("receive an error while write to peer: %s.", addr), err, true),
				}
			}
		}
	}
}

func (sp *ServerPeer) handleMessage(msg interface{}) {
	switch msg.(type) {
	case p2p.MsgVersion:
	case p2p.MsgVerack:
	case p2p.MsgWtxidrelay:
	case p2p.MsgPing:
		pp := msg.(p2p.MsgPing)
		pong, _ := p2p.NewPongMsg("mainnet", pp.Nonce)
		sp.outgoingMsgs <- pong
	case p2p.MsgHeaders:
		sp.msgHeaders <- msg.(*p2p.MsgHeaders)
	case p2p.MsgBlock:
		sp.msgBlocks <- msg.(*p2p.MsgBlock)
	default:
		log.Printf("missing handler for msg: %#v\n", msg)
	}
}

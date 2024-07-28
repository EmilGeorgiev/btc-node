package node

import (
	"errors"
	"fmt"
	errors2 "github.com/EmilGeorgiev/btc-node/errors"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type ServerPeer struct {
	msgHandlersManager    StartStop
	peerSync              StartStop
	networkMessageHandler NetworkMessageHandler
	peer                  p2p.Peer
	outgoingMsgs          chan *p2p.Message
	errors                chan<- PeerErr
	isStarted             atomic.Bool
	isSyncStarted         atomic.Bool
	network               string

	msgHeaders chan<- *p2p.MsgHeaders
	msgBlocks  chan<- *p2p.MsgBlock

	stop chan struct{}
	wg   sync.WaitGroup
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
		stop:                  make(chan struct{}, 1),
	}
}

func (sp *ServerPeer) Start() {
	log.Println("Start server peer")
	if sp.isStarted.Load() {
		return
	}
	log.Println("Start server peer 111111")
	sp.isStarted.Store(true)
	log.Println("Start server peer 2222222")
	sp.msgHandlersManager.Start()
	log.Println("Start server peer 33333333")
	sp.wg.Add(2)
	go sp.handleIncomingMsgs(&sp.wg)
	log.Println("Start server peer 4444444")
	go sp.handOutgoingMsgs(&sp.wg)
	log.Println("Start server peer 5555")
	sp.peerSync.Start()
	log.Println("Start server peer 666666")
}

//func (sp *ServerPeer) Sync() {
//	if !sp.isStarted.Load() {
//		log.Println("server peer not started")
//		return
//	}
//	if sp.isSyncStarted.Load() {
//		return
//	}
//	sp.isSyncStarted.Store(true)
//	sp.peerSync.Start()
//}

//func (sp *ServerPeer) StopSync() {
//	if !sp.isSyncStarted.Load() {
//		return
//	}
//	sp.isSyncStarted.Store(false)
//	sp.peerSync.Stop()
//}

func (sp *ServerPeer) Stop() {
	log.Println("stop server peer")
	if !sp.isStarted.Load() {
		return
	}
	sp.isStarted.Store(false)
	log.Println("stop server peer 1111111")
	sp.peerSync.Stop()
	close(sp.stop)
	log.Println("stop server peer 222222")

	log.Println("stop server peer 333333")
	sp.msgHandlersManager.Stop()
	log.Println("stop server peer 4444")

	sp.wg.Wait()
	log.Println("stop server peer 5555555")
	err := sp.peer.Connection.Close()
	log.Println("stop server peer 66666: ", err)
}

//func (sp *ServerPeer) GetPeerAddr() string {
//	return sp.peer.Address
//}

type PeerErr struct {
	Peer p2p.Peer
	Err  error
}

func (sp *ServerPeer) handleIncomingMsgs(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Start gorouine that handle incomming messages")
	conn := sp.peer.Connection
	addr := sp.peer.Address
	for {
		select {
		case <-sp.stop:
			log.Println("Stop goroutine that handle incomming messages from peer: ", addr)
			return
		default:
			msg, err := sp.networkMessageHandler.ReadMessage(conn)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}
				fmt.Println("SEND ERRRRRRORRRRRR")
				sp.errors <- PeerErr{
					Peer: sp.peer,
					Err:  errors2.NewE(fmt.Sprintf("receive an error while reading from peer: %s.", addr), err, true),
				}
				fmt.Println("SEND ERRRRRRORRRRRR --- AFTER AFTER")
				go sp.Stop()
				return
			}
			sp.handleMessage(msg)
		}
	}
}

func (sp *ServerPeer) handOutgoingMsgs(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Start goroutine that handle outgoin messags")
	conn := sp.peer.Connection
	addr := sp.peer.Address
	for {
		select {
		case <-sp.stop:
			log.Println("Stop goroutine that handle outgoing msg for peers: ", addr)
			return
		case msg := <-sp.outgoingMsgs:
			fmt.Println("send outgoin message:", msg.MessageHeader.CommandString())
			err := sp.networkMessageHandler.WriteMessage(msg, conn)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}
				sp.errors <- PeerErr{
					Peer: sp.peer,
					Err:  errors2.NewE(fmt.Sprintf("receive an error while write to peer: %s.", addr), err, true),
				}
				go sp.Stop()
				return
			}
		}
	}
}

func (sp *ServerPeer) handleMessage(msg interface{}) {
	switch msg.(type) {
	case *p2p.MsgVersion:
	case *p2p.MsgVerack:
	case *p2p.MsgWtxidrelay:
	case *p2p.MsgPing:
		pp := msg.(*p2p.MsgPing)
		pong, _ := p2p.NewPongMsg("mainnet", pp.Nonce)
		fmt.Println("SEND outgoin msgs ping")
		sp.outgoingMsgs <- pong
		fmt.Println("SEND outgoin msgs ping -- AFTER")
	case *p2p.MsgHeaders:
		fmt.Println("Send msg headers to cahnnel")
		sp.msgHeaders <- msg.(*p2p.MsgHeaders)
		fmt.Println("Send msg headers to cahnnel-- AFTER")
	case *p2p.MsgBlock:
		sp.msgBlocks <- msg.(*p2p.MsgBlock)
	default:
		//log.Printf("missing handler for msg: %#v\n", msg)
	}
}

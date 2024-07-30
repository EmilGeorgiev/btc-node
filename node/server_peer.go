package node

import (
	"errors"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/common"
	errors2 "github.com/EmilGeorgiev/btc-node/errors"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Mode int

const (
	Overview Mode = iota
	Standard
)

type ServerPeer struct {
	msgHandlersManager    MsgHandlersManager
	peerSync              SyncManager
	networkMessageHandler NetworkMessageHandler
	peer                  p2p.Peer
	outgoingMsgs          chan *p2p.Message
	errors                chan<- PeerErr
	network               string
	//mode                  Mode
	mode       atomic.Int64
	msgHeaders chan<- *p2p.MsgHeaders

	msgBlocks chan<- *p2p.MsgBlock
	stop      chan struct{}

	wg sync.WaitGroup

	isStarted atomic.Bool
	//isSyncStarted         atomic.Bool
	//isOverviewStarted      atomic.Bool
}

func NewServerPeer(network string, mhm MsgHandlersManager, ps SyncManager, nmh NetworkMessageHandler, p p2p.Peer,
	out chan *p2p.Message, e chan<- PeerErr, h chan<- *p2p.MsgHeaders, b chan<- *p2p.MsgBlock) *ServerPeer {
	sp := &ServerPeer{
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
	sp.mode.Store(int64(Overview))
	return sp
}

func (sp *ServerPeer) Start() {
	if sp.isStarted.Load() {
		log.Println("ServerPeer is already started.")
		return
	}
	sp.isStarted.Store(true)
	sp.msgHandlersManager.StartOverviewHandlers()
	sp.wg.Add(2)
	go sp.handleIncomingMsgs(&sp.wg)
	go sp.handOutgoingMsgs(&sp.wg)
	log.Println("Start ServerPeer.")
}

func (sp *ServerPeer) GetPeerAddr() string {
	return sp.peer.Address
}

func (sp *ServerPeer) GetChainOverview() (<-chan common.ChainOverview, error) {
	ch := make(chan common.ChainOverview, 1)
	sp.mode.Store(int64(Overview))
	sp.peerSync.StartChainOverview(sp.peer.Address, ch)
	return ch, nil
}

func (sp *ServerPeer) Sync() {
	//if !sp.isStarted.Load() {
	//	log.Println("server peer not started")
	//	return
	//}
	//if sp.isSyncStarted.Load() {
	//	log.Println("sync process with this peer is already started.")
	//	return
	//}
	//sp.isSyncStarted.Store(true)
	sp.mode.Store(int64(Standard))
	sp.peerSync.Start()
}

func (sp *ServerPeer) StopSync() {
	sp.peerSync.Stop()
}

func (sp *ServerPeer) Stop() {
	if !sp.isStarted.Load() {
		log.Println("Can't stop ServerPeer because it is not started.")
		return
	}
	sp.isStarted.Store(false)
	sp.peerSync.Stop()
	close(sp.stop)
	sp.msgHandlersManager.Stop()
	sp.wg.Wait()
	err := sp.peer.Connection.Close()
	log.Println("Stop ServerPeer:", err)
}

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
				sp.errors <- PeerErr{
					Peer: sp.peer,
					Err:  errors2.NewE(fmt.Sprintf("receive an error while reading from peer: %s.", addr), err, true),
				}

				log.Println("server peer stop itself")
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

			if sp.mode.Load() == int64(Overview) && msg.CommandString() != p2p.CmdGetheaders && msg.CommandString() != p2p.CmdPong {
				continue
			}
			log.Println("send outgoin message:", msg.MessageHeader.CommandString())
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
		sp.outgoingMsgs <- pong
	case *p2p.MsgHeaders:
		sp.msgHeaders <- msg.(*p2p.MsgHeaders)
	case *p2p.MsgBlock:
		if sp.mode.Load() == int64(Overview) {
			return
		}
		sp.msgBlocks <- msg.(*p2p.MsgBlock)
	default:
		//log.Printf("missing handler for msg: %#v\n", msg)
	}
}

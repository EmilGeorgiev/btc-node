package node

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgHeadersHandler struct {
	network                  string
	outgoingMsgs             chan<- *p2p.Message
	headers                  <-chan *p2p.MsgHeaders
	expectedStartFromHash    <-chan [32]byte
	syncCompleted            chan<- struct{}
	stop                     chan struct{}
	done                     chan struct{}
	isStarted                atomic.Bool
	notifyForExpectedHeaders chan<- []p2p.BlockHeader
}

func NewMsgHeaderHandler(n string, out chan<- *p2p.Message, h <-chan *p2p.MsgHeaders,
	expectedStartFromHash <-chan [32]byte, syncCompl chan struct{}, notifyForExpectedBlockHeaders chan<- []p2p.BlockHeader) *MsgHeadersHandler {
	return &MsgHeadersHandler{
		network:                  n,
		outgoingMsgs:             out,
		headers:                  h,
		expectedStartFromHash:    expectedStartFromHash,
		syncCompleted:            syncCompl,
		stop:                     make(chan struct{}),
		done:                     make(chan struct{}),
		notifyForExpectedHeaders: notifyForExpectedBlockHeaders,
	}
}

func (mh *MsgHeadersHandler) Start() {
	if mh.isStarted.Load() {
		return
	}
	mh.isStarted.Store(true)
	go mh.handleHeaders()
}

func (mh *MsgHeadersHandler) Stop() {
	if !mh.isStarted.Load() {
		return
	}
	mh.stop <- struct{}{}
	<-mh.done
}

type HeadersFromPeer struct {
	Headers  p2p.MsgHeaders
	PeerAddr string
}

func (mh *MsgHeadersHandler) handleHeaders() {
	fmt.Println("START HEADERS HANDLER")
	var expHeadersToStartFromHash = [32]byte{}
	for {
		select {
		case <-mh.stop:
			mh.done <- struct{}{}
			return
		case expHeadersToStartFromHash = <-mh.expectedStartFromHash:
		case msgH := <-mh.headers: // handle MsgHeaders
			headers := msgH.BlockHeaders
			if len(headers) == 0 {
				log.Println("complete sync")
				mh.syncCompleted <- struct{}{}
				continue
			}
			if headers[0].PrevBlockHash != expHeadersToStartFromHash {
				log.Println("headers start with unexpected hash")
				log.Println("expected:", expHeadersToStartFromHash)
				log.Println("actual:", headers[0].PrevBlockHash)
				continue
			}

			inv := make([]p2p.InvVector, len(headers))
			for i := 0; i < len(msgH.BlockHeaders); i++ {
				inv[i] = p2p.InvVector{Type: 2, Hash: headers[i].PrevBlockHash}
			}

			msgGetdata := p2p.MsgGetData{Count: p2p.VarInt(len(headers)), Inventory: inv}
			msg, _ := p2p.NewMessage(p2p.CmdGetdata, mh.network, msgGetdata)
			fmt.Println("send get data with ", msgGetdata.Count)
			mh.notifyForExpectedHeaders <- msgH.BlockHeaders // notify block handlers what to expect
			mh.outgoingMsgs <- msg
		}
	}
}
func Hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b[:80])
	return sha256.Sum256(firstHash[:])
}

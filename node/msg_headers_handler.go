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
		stop:                     make(chan struct{}, 1000),
		done:                     make(chan struct{}, 1000),
		notifyForExpectedHeaders: notifyForExpectedBlockHeaders,
	}
}

func (mh *MsgHeadersHandler) Start() {
	log.Println("Start MsgHEADERS handler")
	if mh.isStarted.Load() {
		return
	}
	log.Println("Start MsgHEADERS handler 111111111")
	mh.isStarted.Store(true)
	log.Println("Start MsgHEADERS handler 222222")
	go mh.handleHeaders()
	log.Println("Start MsgHEADERS handler 333333333")
}

func (mh *MsgHeadersHandler) Stop() {
	log.Println("Stop MsgHEADERS handler")
	if !mh.isStarted.Load() {
		return
	}
	log.Println("Stop MsgHEADERS handler 1111111")
	mh.stop <- struct{}{}
	log.Println("Stop MsgHEADERS handler 2222222")
	<-mh.done
	log.Println("Stop MsgHEADERS handler 333333333")
}

type HeadersFromPeer struct {
	Headers  p2p.MsgHeaders
	PeerAddr string
}

func (mh *MsgHeadersHandler) handleHeaders() {
	fmt.Println("START HEADERS HANDLER")
	//var expHeadersToStartFromHash = [32]byte{}
	for {
		select {
		case <-mh.stop:
			mh.done <- struct{}{}
			return
		case <-mh.expectedStartFromHash:
		case msgH := <-mh.headers: // handle MsgHeaders
			headers := msgH.BlockHeaders
			if len(headers) == 0 {
				log.Println("complete sync")
				mh.syncCompleted <- struct{}{}
				continue
			}

			if !ValidateChain(msgH.BlockHeaders) {
				continue
			}

			inv := make([]p2p.InvVector, len(headers))
			for i := 0; i < len(msgH.BlockHeaders); i++ {
				inv[i] = p2p.InvVector{Type: 2, Hash: headers[i].PrevBlockHash}
			}

			msgGetdata := p2p.MsgGetData{Count: p2p.VarInt(len(headers)), Inventory: inv}
			msg, _ := p2p.NewMessage(p2p.CmdGetdata, mh.network, msgGetdata)
			fmt.Println("Send Get Data With ", msgGetdata.Count)
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

func ValidateChain(headers []p2p.BlockHeader) bool {
	for i := 1; i < len(headers); i++ {
		if headers[i].PrevBlockHash != Hash(headers[i-1]) {
			log.Println("block's previous block hash is different")
			return false
		}
		if !blockHashLessThanTargetDifficulty(&headers[i]) {
			hash := Hash(headers[i])
			log.Println("block hash is greather than target difficulty:")
			log.Printf("Hash %x\n", p2p.Reverse(hash[:]))
			log.Println("Bits:", headers[i].Bits)
			return false
		}
	}
	return false
}

package node

import (
	"crypto/sha256"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/sync"
	"log"
	"math/big"
	"sync/atomic"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgHeadersHandler struct {
	network               string
	outgoingMsgs          chan<- *p2p.Message
	headers               <-chan *p2p.MsgHeaders
	expectedStartFromHash <-chan [32]byte
	syncCompleted         chan<- struct{}
	stop                  chan struct{}
	done                  chan struct{}
	isStarted             atomic.Bool
	headersOverviews      chan<- sync.HeadersOverview //[]p2p.BlockHeader
}

func NewMsgHeaderHandler(n string, out chan<- *p2p.Message, h <-chan *p2p.MsgHeaders,
	expectedStartFromHash <-chan [32]byte, syncCompl chan struct{}, headersOverviews chan<- sync.HeadersOverview) *MsgHeadersHandler {
	return &MsgHeadersHandler{
		network:               n,
		outgoingMsgs:          out,
		headers:               h,
		expectedStartFromHash: expectedStartFromHash,
		syncCompleted:         syncCompl,
		stop:                  make(chan struct{}, 1000),
		done:                  make(chan struct{}, 1000),
		headersOverviews:      headersOverviews,
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
	expPrevBlockHash := zero
	for {
		select {
		case <-mh.stop:
			mh.done <- struct{}{}
			return
		case expPrevBlockHash = <-mh.expectedStartFromHash:
		case msgH := <-mh.headers: // handle MsgHeaders
			headers := msgH.BlockHeaders
			if len(headers) == 0 {
				log.Println("complete sync")
				mh.syncCompleted <- struct{}{}
				continue
			}

			if expPrevBlockHash != headers[0].PrevBlockHash {
				log.Println("The current Headers are not requested and will be scipped")
				continue
			}

			cumulPoW, isValid := ValidateChain(msgH.BlockHeaders)
			if !isValid {
				lastBlockHash := Hash(msgH.BlockHeaders[len(msgH.BlockHeaders)-1])
				mh.headersOverviews <- sync.HeadersOverview{
					LastBlockHash: lastBlockHash,
					HeadersCount:  int64(len(msgH.BlockHeaders)),
					CumulativePoW: cumulPoW,
					IsValid:       true,
				}
				continue
			}

			inv := make([]p2p.InvVector, len(headers))
			for i := 0; i < len(msgH.BlockHeaders); i++ {
				inv[i] = p2p.InvVector{Type: 2, Hash: headers[i].PrevBlockHash}
			}

			msgGetdata := p2p.MsgGetData{Count: p2p.VarInt(len(headers)), Inventory: inv}
			msg, _ := p2p.NewMessage(p2p.CmdGetdata, mh.network, msgGetdata)
			fmt.Println("Send Get Data With ", msgGetdata.Count)
			lastBlockHash := Hash(msgH.BlockHeaders[len(msgH.BlockHeaders)-1])
			mh.headersOverviews <- sync.HeadersOverview{
				LastBlockHash: lastBlockHash,
				HeadersCount:  int64(len(msgH.BlockHeaders)),
				CumulativePoW: cumulPoW,
				IsValid:       true,
			}
			// notify block handlers what to expect
			mh.outgoingMsgs <- msg
		}
	}
}
func Hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b[:80])
	return sha256.Sum256(firstHash[:])
}

func ValidateChain(headers []p2p.BlockHeader) (*big.Int, bool) {
	for i := 1; i < len(headers); i++ {
		if headers[i].PrevBlockHash != Hash(headers[i-1]) {
			log.Println("block's previous block hash is different")
			return nil, false
		}
		if !blockHashLessThanTargetDifficulty(&headers[i]) {
			hash := Hash(headers[i])
			log.Println("block hash is greather than target difficulty:")
			log.Printf("Hash %x\n", p2p.Reverse(hash[:]))
			log.Println("Bits:", headers[i].Bits)
			return nil, false
		}
	}
	return false
}

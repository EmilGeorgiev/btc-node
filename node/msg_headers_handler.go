package node

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"sync/atomic"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
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
	headersOverviews      chan<- sync.RequestedHeaders
}

func NewMsgHeaderHandler(n string, out chan<- *p2p.Message, h <-chan *p2p.MsgHeaders,
	expectedStartFromHash <-chan [32]byte, syncCompl chan struct{}, headersOverviews chan<- sync.RequestedHeaders) *MsgHeadersHandler {
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
	if mh.isStarted.Load() {
		log.Println("MsgHeadersHandler is already started.")
		return
	}
	mh.isStarted.Store(true)
	go mh.handleHeaders()
	log.Println("Start MsgHeadersHandler.")
}

func (mh *MsgHeadersHandler) Stop() {
	if !mh.isStarted.Load() {
		log.Println("MsgHeadersHandler is already started.")
		return
	}
	mh.stop <- struct{}{}
	<-mh.done
	log.Println("Stop MsgHeadersHandler")
}

type HeadersFromPeer struct {
	Headers  p2p.MsgHeaders
	PeerAddr string
}

func (mh *MsgHeadersHandler) handleHeaders() {
	fmt.Println("START HEADERS HANDLER")
	expPrevBlockHash := sync.GenesisBlockHash
	for {
		select {
		case <-mh.stop:
			mh.done <- struct{}{}
			return
		case expPrevBlockHash = <-mh.expectedStartFromHash:
			log.Printf("set expPrevBlockhash: %x\n", p2p.Reverse(expPrevBlockHash))
		case msgH := <-mh.headers: // handle MsgHeaders
			headers := msgH.BlockHeaders
			if len(headers) == 0 {
				log.Println("complete sync")
				//mh.syncCompleted <- struct{}{}
				mh.headersOverviews <- sync.RequestedHeaders{IsValid: true}
				continue
			}

			if expPrevBlockHash != headers[0].PrevBlockHash {
				log.Println("The current Headers are not requested and will be scipped")
				log.Printf("expected prev block hash: %x\n", p2p.Reverse(expPrevBlockHash))
				log.Printf("actual prev block hash: %x\n", p2p.Reverse(headers[0].PrevBlockHash))
				continue
			}

			cumulPoW, isValid := ValidateChain(msgH.BlockHeaders)
			if !isValid {
				lastBlockHash := Hash(msgH.BlockHeaders[len(msgH.BlockHeaders)-1])
				log.Printf("headers chain is not valid ; %x\n", p2p.Reverse(lastBlockHash))
				mh.headersOverviews <- sync.RequestedHeaders{
					CumulativePoW: cumulPoW,
					IsValid:       false,
				}
				continue
			}

			inv := make([]p2p.InvVector, len(headers))
			for i := 0; i < len(msgH.BlockHeaders); i++ {
				inv[i] = p2p.InvVector{Type: 2, Hash: Hash(headers[i])}
			}

			msgGetdata := p2p.MsgGetData{Count: p2p.VarInt(len(headers)), Inventory: inv}
			msg, _ := p2p.NewMessage(p2p.CmdGetdata, mh.network, msgGetdata)
			log.Println("Send Get Data With ", msgGetdata.Count)
			mh.headersOverviews <- sync.RequestedHeaders{
				BlockHeaders:  headers,
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
	cumulPoW := big.NewInt(0)
	for i := 1; i < len(headers); i++ {
		if headers[i].PrevBlockHash != Hash(headers[i-1]) {
			h := headers[i].PrevBlockHash
			log.Printf("block's previous block hash is different. prev block hash: %x\n", p2p.Reverse(h))
			return nil, false
		}
		if !blockHashLessThanTargetDifficulty(&headers[i]) {
			log.Println("block hash is greather than target difficulty:")
			log.Printf("Hash %x\n", p2p.Reverse(Hash(headers[i])))
			log.Println("Bits:", headers[i].Bits)
			return nil, false
		}

		target := BitsToTarget(headers[i].Bits)
		currentPoW := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil)
		currentPoW.Div(currentPoW, target)
		cumulPoW.Add(cumulPoW, currentPoW)
	}
	return cumulPoW, true
}

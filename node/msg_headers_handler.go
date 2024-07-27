package node

import (
	"crypto/sha256"
	"fmt"
	"sync/atomic"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgHeadersHandler struct {
	network             string
	outgoingMsgs        chan<- *p2p.Message
	headers             <-chan *p2p.MsgHeaders
	expectedBlockHashes <-chan [32]byte
	syncCompleted       chan<- struct{}
	stop                chan struct{}
	done                chan struct{}
	isStarted           atomic.Bool
}

func NewMsgHeaderHandler(n string, out chan<- *p2p.Message, h <-chan *p2p.MsgHeaders,
	expectedBlockHashes <-chan [32]byte, sf chan struct{}) *MsgHeadersHandler {
	return &MsgHeadersHandler{
		network:             n,
		outgoingMsgs:        out,
		headers:             h,
		expectedBlockHashes: expectedBlockHashes,
		syncCompleted:       sf,
		stop:                make(chan struct{}),
		done:                make(chan struct{}),
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
	var expectPrevBlockHash = [32]byte{}
	for {
		select {
		case <-mh.stop:
			mh.done <- struct{}{}
			return
		case expectPrevBlockHash = <-mh.expectedBlockHashes:
		case msgH := <-mh.headers: // handle MsgHeaders
			headers := msgH.BlockHeaders
			if len(headers) == 0 {
				mh.syncCompleted <- struct{}{}
				continue
			}
			fmt.Println("check prev block vs expected block hash")
			fmt.Println("actaul prev block hash: ", headers[0].PrevBlockHash)
			fmt.Println("expected: ", expectPrevBlockHash)
			//if headers[0].PrevBlockHash != expectPrevBlockHash {
			//	continue
			//}

			fmt.Println("Build get data")
			inv := make([]p2p.InvVector, len(headers))
			for i := 0; i < len(msgH.BlockHeaders); i++ {
				inv[i] = p2p.InvVector{Type: 2, Hash: headers[i].PrevBlockHash}
			}

			msgGetdata := p2p.MsgGetData{Count: p2p.VarInt(len(headers)), Inventory: inv}
			msg, _ := p2p.NewMessage(p2p.CmdGetdata, mh.network, msgGetdata)
			fmt.Println("send get data with ", msgGetdata.Count)
			mh.outgoingMsgs <- msg
		}
	}
}
func Hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b)
	return sha256.Sum256(firstHash[:])
}

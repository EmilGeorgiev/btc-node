package node

import (
	"crypto/sha256"
	"fmt"

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
}

func NewMsgHeaderHandler(n string, out chan<- *p2p.Message, h <-chan *p2p.MsgHeaders,
	expectedBlockHashes <-chan [32]byte, sf chan struct{}) MsgHeadersHandler {
	return MsgHeadersHandler{
		network:             n,
		outgoingMsgs:        out,
		headers:             h,
		expectedBlockHashes: expectedBlockHashes,
		syncCompleted:       sf,
		stop:                make(chan struct{}),
		done:                make(chan struct{}),
	}
}

func (mh MsgHeadersHandler) Start() {
	go mh.handleHeaders()
}

func (mh MsgHeadersHandler) Stop() {
	close(mh.stop)
	<-mh.done
}

type HeadersFromPeer struct {
	Headers  p2p.MsgHeaders
	PeerAddr string
}

func (mh MsgHeadersHandler) handleHeaders() {
	fmt.Println("START HEADERS HANDLER")
	var expectPrevBlockHash = [32]byte{}
	for {
		select {
		case <-mh.stop:
			close(mh.done)
			return
		case expectPrevBlockHash = <-mh.expectedBlockHashes:
			fmt.Println("expect received block hashes")
		case msgH := <-mh.headers: // handle MsgHeaders
			fmt.Println("YESSSSSSSSSSSSSSSSSSSSSSSS")
			fmt.Println("YESSSSSSSSSSSSSSSSSSSSSSSS")
			fmt.Println("YESSSSSSSSSSSSSSSSSSSSSSSS")
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
			fmt.Println("send put data")
			mh.outgoingMsgs <- msg
			fmt.Println("after channel get data")
		}
	}
}
func Hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b)
	return sha256.Sum256(firstHash[:])
}

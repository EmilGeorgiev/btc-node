package sync

import (
	"crypto/sha256"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgHeadersHandler struct {
	network     string
	msgSender   MsgSender
	headers     <-chan p2p.MsgHeaders
	blockHashes <-chan [32]byte
	stop        <-chan struct{}
}

func NewMsgHeaderHandler(n string, ms MsgSender, h <-chan p2p.MsgHeaders, b <-chan [32]byte, s <-chan struct{}) MsgHeadersHandler {
	return MsgHeadersHandler{
		network:     n,
		msgSender:   ms,
		headers:     h,
		blockHashes: b,
		stop:        s,
	}
}

func (mh MsgHeadersHandler) HandleMsgHeaders() {
	go mh.handleHeaders()
}

func (mh MsgHeadersHandler) handleHeaders() {
	var expectBlockHeadersStartedFromHash [32]byte
	for {
		select {
		case <-mh.stop:
			return
		case expectBlockHeadersStartedFromHash = <-mh.blockHashes:
		case msgH := <-mh.headers: // handle MsgHeaders
			if (len(msgH.BlockHeaders) == 0) || (msgH.BlockHeaders[0].PrevBlockHash != expectBlockHeadersStartedFromHash) {
				continue
			}

			inv := make([]p2p.InvVector, len(msgH.BlockHeaders))
			for i := 0; i < len(msgH.BlockHeaders); i++ {
				inv[i] = p2p.InvVector{Type: 2, Hash: Hash(msgH.BlockHeaders[i])}
			}

			msgGetdata := p2p.MsgGetData{Count: p2p.VarInt(len(msgH.BlockHeaders)), Inventory: inv}
			msg, _ := p2p.NewMessage(p2p.CmdGetdata, mh.network, msgGetdata)
			if err := mh.msgSender.SendMsg(*msg); err != nil {
				expectBlockHeadersStartedFromHash = [32]byte{}
			}
		}
	}
}
func Hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b)
	return sha256.Sum256(firstHash[:])
}

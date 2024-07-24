package sync_test

import (
	"crypto/sha256"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestStartSync(t *testing.T) {
	messages := make(chan p2p.Message, 1000)

	ctrl := gomock.NewController(t)
	blockRepo := sync.NewMockBlockRepository(ctrl)
	msgSender := sync.NewMockMsgSender(ctrl)

	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
	lastBlockLocally := newMsgBlock(prevBlockHash)

	msgGetHeather, _ := p2p.NewMsgGetHeader("mainnet", 1, lastBlockLocally.GetHash(), [32]byte{0})

	blockRepo.EXPECT().GetLast().Return(lastBlockLocally)
	msgSender.EXPECT().SendMsg(msgGetHeather).Return(nil)

	bh1 := newBlockHeader(hash(lastBlockLocally.BlockHeader))
	bh2 := newBlockHeader(hash(bh1))
	bh3 := newBlockHeader(hash(bh2))
	headers := p2p.MsgHeaders{
		Count:        3,
		BlockHeaders: []p2p.BlockHeader{bh1, bh2, bh3},
	}

	msg, _ := p2p.NewMessage("headers", "mainnet", headers)
	messages <- *msg

	m.Start()

}

func newMsgBlock(prevBlockHash [32]byte) p2p.MsgBlock {

	return p2p.MsgBlock{
		BlockHeader: p2p.BlockHeader{
			Version:       1,
			PrevBlockHash: prevBlockHash,
			MerkleRoot:    [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A},
			Timestamp:     1721836804,
			Bits:          1721836804,
			Nonce:         1721836804,
			TxnCount:      p2p.VarInt(1),
		},
	}
}

func newBlockHeader(prevBlockHash [32]byte) p2p.BlockHeader {
	return p2p.BlockHeader{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A},
		Timestamp:     1721836804,
		Bits:          1721836804,
		Nonce:         1721836804,
		TxnCount:      p2p.VarInt(1),
	}
}

func hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b)
	return sha256.Sum256(firstHash[:])
}

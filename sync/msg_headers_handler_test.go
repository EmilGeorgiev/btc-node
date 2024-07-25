package sync_test

import (
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestHandleMsgHeaders_HappyPath(t *testing.T) {
	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bh1 := newBlockHeader(prevBlockHash)
	bh2 := newBlockHeader(sync.Hash(bh1))
	bh3 := newBlockHeader(sync.Hash(bh2))
	blockHeaders := []p2p.BlockHeader{bh1, bh2, bh3}
	msgHeaders := p2p.MsgHeaders{Count: 3, BlockHeaders: blockHeaders}

	msggetdata := p2p.MsgGetData{
		Count: 3,
		Inventory: []p2p.InvVector{
			{2, sync.Hash(bh1)},
			{2, sync.Hash(bh2)},
			{2, sync.Hash(bh3)},
		},
	}

	msgGetData, _ := p2p.NewMessage(p2p.CmdGetdata, "mainnet", msggetdata)

	ctrl := gomock.NewController(t)
	msgSender := sync.NewMockMsgSender(ctrl)
	msgSender.EXPECT().SendMsg(*msgGetData).Return(nil).Times(1)

	stop := make(chan struct{})
	expectedBlockHashes := make(chan [32]byte)
	msgHeadersCh := make(chan p2p.MsgHeaders)
	headersHandler := sync.NewMsgHeaderHandler("mainnet", msgSender, msgHeadersCh, expectedBlockHashes, stop)
	headersHandler.HandleMsgHeaders()

	expectedBlockHashes <- prevBlockHash
	msgHeadersCh <- msgHeaders

	headersHandler.Stop()
}

func TestHandleMsgHeaders_WhenMsgHeadersHasZeroBlockHeaders(t *testing.T) {
	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	blockHeaders := []p2p.BlockHeader{}
	msgHeaders := p2p.MsgHeaders{Count: 0, BlockHeaders: blockHeaders}

	ctrl := gomock.NewController(t)
	msgSender := sync.NewMockMsgSender(ctrl)
	msgSender.EXPECT().SendMsg(gomock.Any()).Times(0)

	stop := make(chan struct{})
	expectedBlockHashes := make(chan [32]byte)
	msgHeadersCh := make(chan p2p.MsgHeaders)
	headersHandler := sync.NewMsgHeaderHandler("mainnet", msgSender, msgHeadersCh, expectedBlockHashes, stop)
	headersHandler.HandleMsgHeaders()

	expectedBlockHashes <- prevBlockHash
	msgHeadersCh <- msgHeaders

	headersHandler.Stop()
}

func TestHandleMsgHeaders_WhenMsgHeadersContainsUnwantedBlockHeaders(t *testing.T) {
	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bh1 := newBlockHeader([32]byte{0})
	bh2 := newBlockHeader(sync.Hash(bh1))
	bh3 := newBlockHeader(sync.Hash(bh2))
	blockHeaders := []p2p.BlockHeader{bh1, bh2, bh3}
	msgHeaders := p2p.MsgHeaders{Count: 3, BlockHeaders: blockHeaders}

	ctrl := gomock.NewController(t)
	msgSender := sync.NewMockMsgSender(ctrl)
	msgSender.EXPECT().SendMsg(gomock.Any()).Times(0)

	stop := make(chan struct{})
	expectedBlockHashes := make(chan [32]byte)
	msgHeadersCh := make(chan p2p.MsgHeaders)
	headersHandler := sync.NewMsgHeaderHandler("mainnet", msgSender, msgHeadersCh, expectedBlockHashes, stop)
	headersHandler.HandleMsgHeaders()

	expectedBlockHashes <- prevBlockHash
	msgHeadersCh <- msgHeaders

	headersHandler.Stop()
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

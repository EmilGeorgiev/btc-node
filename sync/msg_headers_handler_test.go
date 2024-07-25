package sync_test

import (
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"testing"
)

func TestChanSync_RequestHeatherFromLastBlock(t *testing.T) {
	//prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
	//lastBlockLocally := newMsgBlock(prevBlockHash)
	//
	//payload := &p2p.MsgGetHeader{Version: p2p.Version, HashCount: 1, StartBlock: lastBlockLocally.GetHash(), StopBlock: [32]byte{0}}
	//msgGetHeather, _ := p2p.NewMessage(p2p.CmdGetheaders, "mainnet", payload)
	//
	//ctrl := gomock.NewController(t)
	//blockRepo := NewMockBlockRepository(ctrl)
	//msgSender := NewMockMsgSender(ctrl)
	//blockRepo.EXPECT().GetLast().Return(lastBlockLocally, nil)
	//msgSender.EXPECT().SendMsg(*msgGetHeather).Return(nil)
	//
	//chSync := NewChainSync("mainnet", blockRepo, msgSender, nil, nil)
	//err := chSync.RequestHeadersFromLastBlock()
	//require.NoError(t, err)
	//
	//got := <-chSync.expectedHeaders
	//require.Equal(t, *payload, got)
}

func TestChanSync_HandleMsgHeaders(t *testing.T) {
	//prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
	//lastBlockLocally := newBlockHeader(prevBlockHash)
	//bh1 := newBlockHeader(hash(lastBlockLocally))
	//bh2 := newBlockHeader(hash(bh1))
	//bh3 := newBlockHeader(hash(bh2))
	//headers := p2p.MsgHeaders{Count: 3, BlockHeaders: []p2p.BlockHeader{bh1, bh2, bh3}}
	//
	//headersCh := make(<-chan p2p.MsgHeaders)
	//
	//chSync := NewChainSync("", nil, nil, headersCh, nil)
	//go chSync.HandleMsgHeaders()
	//
	//headersCh <- headers
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

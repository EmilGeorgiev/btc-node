package node_test

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/node"
	"github.com/EmilGeorgiev/btc-node/sync"
	"testing"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHandleMsgBlocks_HappyPath(t *testing.T) {
	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bl1 := newMsgBlock(blockHash)
	bl2 := newMsgBlock(bl1.GetHash())

	fmt.Printf("1: %x\n", p2p.Reverse(bl1.GetHash()))
	fmt.Printf("2: %x\n", p2p.Reverse(bl2.GetHash()))

	ctrl := gomock.NewController(t)
	blockValidator := node.NewMockValidator(ctrl)
	blockValidator.EXPECT().Validate(&bl1).Return(nil).Times(1)
	blockValidator.EXPECT().Validate(&bl2).Return(nil).Times(1)
	blockRepo := node.NewMockBlockRepository(ctrl)
	blockRepo.EXPECT().Save(bl1).Return(nil).Times(1)
	blockRepo.EXPECT().Save(bl2).Return(nil).Times(1)

	blocks := make(chan *p2p.MsgBlock)
	expectedHeaders := make(chan sync.RequestedHeaders)
	processed := make(chan sync.RequestedHeaders)
	msgBlockHandle := node.NewMsgBlockHandler(blockRepo, blockValidator, blocks, processed, expectedHeaders)
	msgBlockHandle.Start()

	expHead := sync.RequestedHeaders{BlockHeaders: []p2p.BlockHeader{bl1.BlockHeader, bl2.BlockHeader}}

	expectedHeaders <- expHead

	blocks <- &bl1
	blocks <- &bl2
	actual := <-processed
	require.Equal(t, expHead, actual)

	msgBlockHandle.Stop()
}

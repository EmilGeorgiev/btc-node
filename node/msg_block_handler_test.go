package node_test

import (
	"errors"
	"github.com/EmilGeorgiev/btc-node/node"
	"testing"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHandleMsgBlocks_HappyPath(t *testing.T) {
	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bl1 := newMsgBlock(blockHash)
	bl2 := newMsgBlock(bl1.GetHash())

	ctrl := gomock.NewController(t)
	blockValidator := sync.NewMockBlockValidator(ctrl)
	blockValidator.EXPECT().Validate(bl1).Return(nil).Times(1)
	blockValidator.EXPECT().Validate(bl2).Return(nil).Times(1)
	blockRepo := sync.NewMockBlockRepository(ctrl)
	blockRepo.EXPECT().Save(bl1).Return(nil).Times(1)
	blockRepo.EXPECT().Save(bl2).Return(nil).Times(1)

	stop := make(chan struct{})
	blocks := make(chan p2p.MsgBlock)
	notify := make(chan p2p.MsgBlock)
	msgBlockHandle := node.NewMsgBlockHandler(blockRepo, blockValidator, blocks, stop, notify)
	msgBlockHandle.HandleMsgBlock()

	blocks <- bl1
	actual := <-notify
	require.Equal(t, bl1, actual)

	blocks <- bl2
	actual = <-notify
	require.Equal(t, bl2, actual)

	msgBlockHandle.Stop()
}

func TestHandleMsgBlocks_WhenValidateFail(t *testing.T) {
	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bl1 := newMsgBlock(blockHash)

	ctrl := gomock.NewController(t)
	blockValidator := sync.NewMockBlockValidator(ctrl)
	blockValidator.EXPECT().Validate(bl1).Return(errors.New("err")).Times(1)
	blockRepo := sync.NewMockBlockRepository(ctrl)
	blockRepo.EXPECT().Save(gomock.Any()).Times(0)

	stop := make(chan struct{})
	blocks := make(chan p2p.MsgBlock)
	notify := make(chan p2p.MsgBlock, 1)
	msgBlockHandle := node.NewMsgBlockHandler(blockRepo, blockValidator, blocks, stop, notify)
	msgBlockHandle.HandleMsgBlock()

	blocks <- bl1
	actual := len(notify)
	require.Equal(t, 0, actual)

	msgBlockHandle.Stop()
}

func TestHandleMsgBlocks_WhenSaveFail(t *testing.T) {
	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bl1 := newMsgBlock(blockHash)

	ctrl := gomock.NewController(t)
	blockValidator := sync.NewMockBlockValidator(ctrl)
	blockValidator.EXPECT().Validate(bl1).Return(nil).Times(1)
	blockRepo := sync.NewMockBlockRepository(ctrl)
	blockRepo.EXPECT().Save(bl1).Return(errors.New("err")).Times(1)

	stop := make(chan struct{})
	blocks := make(chan p2p.MsgBlock)
	notify := make(chan p2p.MsgBlock)
	msgBlockHandle := node.NewMsgBlockHandler(blockRepo, blockValidator, blocks, stop, notify)
	msgBlockHandle.HandleMsgBlock()

	blocks <- bl1
	actual := len(notify)
	require.Equal(t, 0, actual)

	msgBlockHandle.Stop()
}

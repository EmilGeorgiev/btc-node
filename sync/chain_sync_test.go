package sync_test

import (
	"testing"
	"time"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestChainSync_Start_AndStop(t *testing.T) {
	lastBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	peerAddr := "8.8.8.8/32"

	ctrl := gomock.NewController(t)
	node := sync.NewMockNode(ctrl)
	node.EXPECT().GetPeerAddress().Return(peerAddr).Times(1)
	headerRequester := sync.NewMockHeaderRequester(ctrl)
	headerRequester.EXPECT().RequestHeadersFromLastBlock(peerAddr).Return(lastBlockHash, nil)
	headersHandler := sync.NewMockHeadersHandler(ctrl)
	headersHandler.EXPECT().HandleMsgHeaders()
	msgBlockHandler := sync.NewMockBlockHandler(ctrl)
	msgBlockHandler.EXPECT().HandleBlockMessages()

	startFromHashes := make(chan [32]byte)
	processesBlocks := make(chan p2p.MsgBlock)
	chs := sync.NewChainSync(headerRequester, headersHandler, msgBlockHandler, node, 10*time.Minute, startFromHashes, processesBlocks)
	chs.Start()

	actual := <-startFromHashes
	chs.Stop()
	require.Equal(t, lastBlockHash, actual)
}

func TestChainSync_Start_WhenReceiveProcessedBlocks(t *testing.T) {
	lastBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
	peerAddr := "8.8.8.8/32"

	ctrl := gomock.NewController(t)
	node := sync.NewMockNode(ctrl)
	node.EXPECT().GetPeerAddress().Return(peerAddr).Times(1)
	headerRequester := sync.NewMockHeaderRequester(ctrl)
	headerRequester.EXPECT().RequestHeadersFromLastBlock(peerAddr).Return(lastBlockHash, nil).MaxTimes(1)
	headersHandler := sync.NewMockHeadersHandler(ctrl)
	headersHandler.EXPECT().HandleMsgHeaders().MaxTimes(1)
	msgBlockHandler := sync.NewMockBlockHandler(ctrl)
	msgBlockHandler.EXPECT().HandleBlockMessages().MaxTimes(1)

	startFromHashes := make(chan [32]byte, 1)
	processesBlocks := make(chan p2p.MsgBlock)
	chs := sync.NewChainSync(headerRequester, headersHandler, msgBlockHandler, node, 10*time.Millisecond, startFromHashes, processesBlocks)
	chs.Start()

	actual := <-startFromHashes
	// ChainSync will wait 10 milliseconds (we set this in the constructor above), and if it doesn't receive any
	// notification for processed blocks, it will trigger a new sync iteration (meaning that will request MsgHeaders again).
	// By setting this timer to 20 milliseconds we will prove that during this time the new iteration will not be
	// started if we notify ChainSync for processed blocks in time every 2 milliseconds
	timer := time.NewTimer(20 * time.Millisecond)

Loop:
	for {
		select {
		case <-timer.C:
			break Loop
		default:
			processesBlocks <- p2p.MsgBlock{}
			time.Sleep(2 * time.Millisecond)
		}
	}

	chs.Stop()
	require.Equal(t, lastBlockHash, actual)
}

func TestChainSync_Start_WhenNoProcessedBlockAreReceived(t *testing.T) {
	lastBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
	peerAddr := "8.8.8.8/32"

	ctrl := gomock.NewController(t)
	node := sync.NewMockNode(ctrl)
	node.EXPECT().GetPeerAddress().Return(peerAddr).Times(1)
	headerRequester := sync.NewMockHeaderRequester(ctrl)
	headerRequester.EXPECT().RequestHeadersFromLastBlock(peerAddr).Return(lastBlockHash, nil).AnyTimes()
	headersHandler := sync.NewMockHeadersHandler(ctrl)
	headersHandler.EXPECT().HandleMsgHeaders().AnyTimes()
	msgBlockHandler := sync.NewMockBlockHandler(ctrl)
	msgBlockHandler.EXPECT().HandleBlockMessages().AnyTimes()

	startFromHashes := make(chan [32]byte, 100)
	processesBlocks := make(chan p2p.MsgBlock)
	chs := sync.NewChainSync(headerRequester, headersHandler, msgBlockHandler, node, 10*time.Millisecond, startFromHashes, processesBlocks)
	chs.Start()

	time.Sleep(30 * time.Millisecond)
	chs.Stop()

	close(startFromHashes)
	for actual := range startFromHashes {
		require.Equal(t, lastBlockHash, actual)
	}
}

func TestChainSync_Start_WhenRequestHeadersFailed(t *testing.T) {
	lastBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
	peerAddr := "8.8.8.8/32"
	peerAddr2 := "7.7.7.7/32"

	ctrl := gomock.NewController(t)
	node := sync.NewMockNode(ctrl)
	node.EXPECT().GetPeerAddress().Return(peerAddr).Times(1)
	node.EXPECT().GetPeerAddress().Return(peerAddr2).Times(1)
	headerRequester := sync.NewMockHeaderRequester(ctrl)
	headerRequester.EXPECT().RequestHeadersFromLastBlock(peerAddr).Return([32]byte{}, sync.ErrFailedToSendMsgGetHeaders).Times(1)
	headerRequester.EXPECT().RequestHeadersFromLastBlock(peerAddr2).Return(lastBlockHash, nil).Times(1)
	headersHandler := sync.NewMockHeadersHandler(ctrl)
	headersHandler.EXPECT().HandleMsgHeaders().Times(1)
	msgBlockHandler := sync.NewMockBlockHandler(ctrl)
	msgBlockHandler.EXPECT().HandleBlockMessages().Times(1)

	startFromHashes := make(chan [32]byte)
	processesBlocks := make(chan p2p.MsgBlock)
	chs := sync.NewChainSync(headerRequester, headersHandler, msgBlockHandler, node, 10*time.Millisecond, startFromHashes, processesBlocks)
	chs.Start()

	actual := <-startFromHashes
	chs.Stop()
	require.Equal(t, lastBlockHash, actual)
	require.Equal(t, 0, len(startFromHashes))
}

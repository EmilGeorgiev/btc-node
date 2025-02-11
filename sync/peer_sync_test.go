package sync_test

import (
	"github.com/EmilGeorgiev/btc-node/common/testutil"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/node"
	"github.com/EmilGeorgiev/btc-node/sync"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPeerSync_SetUpStartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	headerRequester := sync.NewMockHeaderRequester(ctrl)
	headerRequester.EXPECT().RequestHeadersFromLastBlock().Return(nil).Times(1)

	requestedHeaders := make(chan sync.RequestedHeaders)
	chs := sync.NewPeerSync(headerRequester, 10*time.Minute, requestedHeaders)
	chs.Start()

	require.Equal(t, 0, len(requestedHeaders))
	chs.Stop()
}

func TestPeerSync_Start_WhenReceiveProcessedBlocks(t *testing.T) {
	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bh := testutil.NewBlockHeader(blockHash)

	hash := node.Hash(bh)

	ctrl := gomock.NewController(t)
	headerRequester := sync.NewMockHeaderRequester(ctrl)
	headerRequester.EXPECT().RequestHeadersFromLastBlock().Return(nil).MaxTimes(1)
	headerRequester.EXPECT().RequestHeadersFromBlockHash(hash).Return(nil).AnyTimes()

	requestedHeaders := make(chan sync.RequestedHeaders)
	chs := sync.NewPeerSync(headerRequester, 10*time.Millisecond, requestedHeaders)
	chs.Start()

	// PeerSync will wait 10 milliseconds (we set this in the constructor above), and if it doesn't receive any
	// notification for processed blocks, it will trigger a new sync iteration (meaning that will request MsgHeaders again).
	// By setting this timer to 20 milliseconds we will prove that during this time the new iteration will not be
	// started if we notify PeerSync for processed blocks in time every 2 milliseconds
	timer := time.NewTimer(20 * time.Millisecond)

Loop:
	for {
		select {
		case <-timer.C:
			break Loop
		default:
			requestedHeaders <- sync.RequestedHeaders{
				BlockHeaders: []p2p.BlockHeader{bh},
			}
			time.Sleep(2 * time.Millisecond)
		}
	}

	chs.Stop()
}

package sync_test

//
//import (
//	"github.com/EmilGeorgiev/btc-node/network/p2p"
//	"github.com/EmilGeorgiev/btc-node/sync"
//	"github.com/golang/mock/gomock"
//	"github.com/stretchr/testify/require"
//	"testing"
//	"time"
//)
//
//func TestPeerSync_SetUpStartAndStop(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	headerRequester := sync.NewMockHeaderRequester(ctrl)
//	headerRequester.EXPECT().RequestHeadersFromLastBlock().Return(nil).Times(1)
//	headersHandler := sync.NewMockStartStop(ctrl)
//	headersHandler.EXPECT().Start().Times(1)
//	headersHandler.EXPECT().Stop().Times(1)
//	msgBlockHandler := sync.NewMockStartStop(ctrl)
//	msgBlockHandler.EXPECT().Start().Times(1)
//	msgBlockHandler.EXPECT().Stop().Times(1)
//
//	processesBlocks := make(chan p2p.MsgBlock)
//	chs := sync.NewPeerSync(headerRequester, headersHandler, msgBlockHandler, 10*time.Minute, processesBlocks)
//	chs.SetUp()
//	chs.Start()
//
//	require.Equal(t, 0, len(processesBlocks))
//	chs.Stop()
//}
//
//func TestPeerSync_Start_WhenReceiveProcessedBlocks(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	headerRequester := sync.NewMockHeaderRequester(ctrl)
//	headerRequester.EXPECT().RequestHeadersFromLastBlock().Return(nil).MaxTimes(1)
//	headersHandler := sync.NewMockStartStop(ctrl)
//	headersHandler.EXPECT().Start().Times(1)
//	headersHandler.EXPECT().Stop().Times(1)
//	msgBlockHandler := sync.NewMockStartStop(ctrl)
//	msgBlockHandler.EXPECT().Start().Times(1)
//	msgBlockHandler.EXPECT().Stop().Times(1)
//
//	processesBlocks := make(chan p2p.MsgBlock)
//	chs := sync.NewPeerSync(headerRequester, nil, nil, 10*time.Millisecond, processesBlocks)
//	chs.SetUp()
//	chs.Start()
//
//	// PeerSync will wait 10 milliseconds (we set this in the constructor above), and if it doesn't receive any
//	// notification for processed blocks, it will trigger a new sync iteration (meaning that will request MsgHeaders again).
//	// By setting this timer to 20 milliseconds we will prove that during this time the new iteration will not be
//	// started if we notify PeerSync for processed blocks in time every 2 milliseconds
//	timer := time.NewTimer(20 * time.Millisecond)
//
//Loop:
//	for {
//		select {
//		case <-timer.C:
//			break Loop
//		default:
//			processesBlocks <- p2p.MsgBlock{}
//			time.Sleep(2 * time.Millisecond)
//		}
//	}
//
//	chs.Stop()
//}

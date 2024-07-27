package node_test

import (
	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/node"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNode_StartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng := node.NewMockPeerConnectionManager(ctrl)
	newPeerConnMng := func(p2p.Peer, chan node.PeerErr) node.PeerConnectionManager { return peerConnMng }
	addrs := []common.Addr{
		{IP: "127.0.0.1", Port: 5555},
		{IP: "127.0.0.2", Port: 6666},
	}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan node.PeerErr)

	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1"}}
	handshake2 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.2"}}

	handshakeManager := node.NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(1)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[1], "mainnet", "test-agent").Return(handshake2, nil).Times(1)
	peerConnMng.EXPECT().Start().Times(2)
	peerConnMng.EXPECT().Sync()
	peerConnMng.EXPECT().StopSync()
	peerConnMng.EXPECT().Stop().Times(2)

	n, err := node.New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	syncCompleted <- struct{}{}
	n.Stop()
}

func TestNode_WhenNodeHasOneConnectedPeerAndReceivePeerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng := node.NewMockPeerConnectionManager(ctrl)
	newPeerConnMng := func(p2p.Peer, chan node.PeerErr) node.PeerConnectionManager { return peerConnMng }
	addrs := []common.Addr{{IP: "127.0.0.1", Port: 5555}}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan node.PeerErr)
	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}

	handshakeManager := node.NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(2)
	peerConnMng.EXPECT().Start().Times(2)
	peerConnMng.EXPECT().Sync().Times(2)
	peerConnMng.EXPECT().GetPeerAddr().Return("127.0.0.1:5555").AnyTimes()
	peerConnMng.EXPECT().StopSync().Times(2)
	peerConnMng.EXPECT().Stop().Times(2)

	n, err := node.New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	time.Sleep(10 * time.Millisecond) // what for all goroutines to start

	// when an error with the peer occurred, then the peerConManager is stop, sync process is stopped and reconnect again.
	// if we have only one peerConnection manager than the sync process will continue with it.
	peerErrors <- node.PeerErr{
		Peer: p2p.Peer{Address: "127.0.0.1:5555"},
	}

	time.Sleep(2 * time.Second) // wait all goroutines to restart again

	n.Stop()
}

func TestNode_WhenReceivePeerErrFromPeerWhatIsNotInSyncWithTheNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng1 := node.NewMockPeerConnectionManager(ctrl)
	peerConnMng2 := node.NewMockPeerConnectionManager(ctrl)
	isFirstTime := true
	newPeerConnMng := func(p2p.Peer, chan node.PeerErr) node.PeerConnectionManager {
		if isFirstTime {
			isFirstTime = false
			return peerConnMng1
		}
		return peerConnMng2
	}
	addrs := []common.Addr{{IP: "127.0.0.1", Port: 5555}, {IP: "127.0.0.2", Port: 6666}}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan node.PeerErr)
	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}
	handshake2 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.2:6666"}}

	handshakeManager := node.NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(1)
	peerConnMng1.EXPECT().Start().Times(1)
	peerConnMng1.EXPECT().Sync().Times(1)
	peerConnMng1.EXPECT().GetPeerAddr().Return("127.0.0.1:5555").AnyTimes()
	peerConnMng1.EXPECT().StopSync().Times(1)
	peerConnMng1.EXPECT().Stop().Times(1)

	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[1], "mainnet", "test-agent").Return(handshake2, nil).Times(1)
	peerConnMng2.EXPECT().Start().Times(1)
	//peerConnMng2.EXPECT().Sync().Times(1)
	peerConnMng2.EXPECT().GetPeerAddr().Return("127.0.0.1:6666").AnyTimes()
	//peerConnMng2.EXPECT().StopSync().Times(1)
	peerConnMng2.EXPECT().Stop().Times(1)

	n, err := node.New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	time.Sleep(10 * time.Millisecond) // what for all goroutines to start

	// when an error with the peer occurred, then the peerConManager is stop, sync process is stopped and reconnect again.
	// if we have only one peerConnection manager than the sync process will continue with it.
	//peerErrors <- node.PeerErr{
	//	Peer: p2p.Peer{Address: "127.0.0.1:5555"},
	//}

	time.Sleep(2 * time.Second) // wait all goroutines to restart again

	n.Stop()
}

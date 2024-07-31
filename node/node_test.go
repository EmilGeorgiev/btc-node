package node

import (
	"math/big"
	"testing"
	"time"

	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNode_StartAndSelectTheBestChainForSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng1 := NewMockPeerConnectionManager(ctrl)
	peerConnMng2 := NewMockPeerConnectionManager(ctrl)
	isFirstCall := true
	newPeerConnMng := func(p2p.Peer, chan PeerErr) PeerConnectionManager {
		if isFirstCall {
			isFirstCall = false
			return peerConnMng1
		}
		return peerConnMng2
	}
	addrs := []common.Addr{
		{IP: "127.0.0.1", Port: 5555},
		{IP: "127.0.0.2", Port: 6666},
	}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan PeerErr)

	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1"}}
	handshake2 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.2"}}

	chOverveiw1 := make(chan common.ChainOverview)
	chOverveiw2 := make(chan common.ChainOverview)

	handshakeManager := NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(1)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[1], "mainnet", "test-agent").Return(handshake2, nil).Times(1)
	peerConnMng1.EXPECT().Start().Times(1)
	peerConnMng1.EXPECT().GetChainOverview().Return(chOverveiw1, nil)
	peerConnMng1.EXPECT().GetPeerAddr().Return("127.0.0.1:5555").AnyTimes()
	peerConnMng1.EXPECT().Sync()
	peerConnMng1.EXPECT().Stop().Times(1)

	peerConnMng2.EXPECT().Start().Times(1)
	peerConnMng2.EXPECT().GetChainOverview().Return(chOverveiw2, nil)
	peerConnMng2.EXPECT().GetPeerAddr().Return("127.0.0.2:6666").AnyTimes()
	peerConnMng2.EXPECT().Stop().Times(1)

	n, err := New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	chOverveiw1 <- common.ChainOverview{
		Peer:           "127.0.0.1:5555",
		NumberOfBlocks: 123,
		CumulativeWork: big.NewInt(9999),
		IsValid:        true,
	}
	chOverveiw2 <- common.ChainOverview{
		Peer:           "127.0.0.1:6666",
		NumberOfBlocks: 129,
		CumulativeWork: big.NewInt(8888),
		IsValid:        true,
	}

	n.Stop()
}

func TestNode_WhenNodeHasOneConnectedPeerAndReceivePeerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng1 := NewMockPeerConnectionManager(ctrl)
	newPeerConnMng := func(p2p.Peer, chan PeerErr) PeerConnectionManager {
		return peerConnMng1
	}
	addrs := []common.Addr{{IP: "127.0.0.1", Port: 5555}}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan PeerErr)

	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1"}}
	chOverveiw1 := make(chan common.ChainOverview)

	handshakeManager := NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(2)
	peerConnMng1.EXPECT().Start().Times(2)
	peerConnMng1.EXPECT().GetChainOverview().Return(chOverveiw1, nil).Times(2)
	peerConnMng1.EXPECT().GetPeerAddr().Return("127.0.0.1:5555").AnyTimes()
	peerConnMng1.EXPECT().Sync().Times(2)
	peerConnMng1.EXPECT().Stop().Times(1)

	n, err := New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	chO := common.ChainOverview{
		Peer:           "127.0.0.1:5555",
		NumberOfBlocks: 123,
		CumulativeWork: big.NewInt(9999),
		IsValid:        true,
	}

	chOverveiw1 <- chO
	peerErrors <- PeerErr{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}
	chOverveiw1 <- chO

	//time.Sleep(5 * time.Second)

	n.Stop()
}

package node

import (
	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestNode_StartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng := NewMockPeerConnectionManager(ctrl)
	newPeerConnMng := func(p2p.Peer, chan PeerErr) PeerConnectionManager { return peerConnMng }
	addrs := []common.Addr{
		{IP: "127.0.0.1", Port: 5555},
		{IP: "127.0.0.2", Port: 6666},
	}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan PeerErr)

	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1"}}
	handshake2 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.2"}}

	handshakeManager := NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(1)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[1], "mainnet", "test-agent").Return(handshake2, nil).Times(1)
	peerConnMng.EXPECT().Start().Times(2)
	peerConnMng.EXPECT().Sync()
	peerConnMng.EXPECT().StopSync()
	peerConnMng.EXPECT().Stop().Times(2)

	n, err := New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	syncCompleted <- struct{}{}
	n.Stop()
}

func TestNode_WhenNodeHasOneConnectedPeerAndReceivePeerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng := NewMockPeerConnectionManager(ctrl)
	newPeerConnMng := func(p2p.Peer, chan PeerErr) PeerConnectionManager { return peerConnMng }
	addrs := []common.Addr{{IP: "127.0.0.1", Port: 5555}}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan PeerErr)
	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}

	handshakeManager := NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(2)
	peerConnMng.EXPECT().Start().Times(2)
	peerConnMng.EXPECT().Sync().Times(2)
	peerConnMng.EXPECT().GetPeerAddr().Return("127.0.0.1:5555").AnyTimes()
	peerConnMng.EXPECT().StopSync().Times(2)
	peerConnMng.EXPECT().Stop().Times(2)

	n, err := New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	time.Sleep(10 * time.Millisecond) // what for all goroutines to start

	// when an error with the peer occurred, then the peerConManager is stop, sync process is stopped and reconnect again.
	// if we have only one peerConnection manager than the sync process will continue with it.
	peerErrors <- PeerErr{
		Peer: p2p.Peer{Address: "127.0.0.1:5555"},
	}

	time.Sleep(2 * time.Second) // wait all goroutines to restart again

	n.Stop()
}

func TestNode_WhenReceivePeerErrFromPeerThatIsNotInSyncWithTheNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng1 := NewMockPeerConnectionManager(ctrl)
	peerConnMng2 := NewMockPeerConnectionManager(ctrl)
	isFirstTime := true
	newPeerConnMng := func(p2p.Peer, chan PeerErr) PeerConnectionManager {
		if isFirstTime {
			isFirstTime = false
			return peerConnMng1
		}
		return peerConnMng2
	}
	addrs := []common.Addr{{IP: "127.0.0.1", Port: 5555}, {IP: "127.0.0.2", Port: 6666}}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan PeerErr)
	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}
	handshake2 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.2:6666"}}

	handshakeManager := NewMockHandshakeManager(ctrl)
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

	n, err := New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	//time.Sleep(10 * time.Millisecond) // what for all goroutines to start

	// when an error with the peer occurred, then the peerConManager is stop, sync process is stopped and reconnect again.
	// if we have only one peerConnection manager than the sync process will continue with it.
	peerErrors <- PeerErr{Peer: p2p.Peer{Address: "127.0.0.1:6666"}}

	time.Sleep(2 * time.Second) // wait all goroutines to restart again

	n.Stop()
}

func TestNode_WhenReceivePeerErrFromPeerThatIsInSyncWithTheNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	peerConnMng1 := NewMockPeerConnectionManager(ctrl)
	peerConnMng2 := NewMockPeerConnectionManager(ctrl)
	newPeerConnMng := func(p p2p.Peer, err chan PeerErr) PeerConnectionManager {
		if p.Address == "127.0.0.1:5555" {
			return peerConnMng1
		}
		return peerConnMng2
	}
	addrs := []common.Addr{{IP: "127.0.0.1", Port: 5555}, {IP: "127.0.0.2", Port: 6666}}
	syncCompleted := make(chan struct{})
	peerErrors := make(chan PeerErr)
	handshake1 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}
	handshake2 := p2p.Handshake{Peer: p2p.Peer{Address: "127.0.0.2:6666"}}

	handshakeManager := NewMockHandshakeManager(ctrl)
	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[0], "mainnet", "test-agent").Return(handshake1, nil).Times(2)
	peerConnMng1.EXPECT().Start().Times(2)
	peerConnMng1.EXPECT().Sync().Times(1)
	peerConnMng1.EXPECT().GetPeerAddr().Return("127.0.0.1:5555").AnyTimes()
	peerConnMng1.EXPECT().StopSync().Times(1)
	peerConnMng1.EXPECT().Stop().Times(2)

	handshakeManager.EXPECT().CreateOutgoingHandshake(addrs[1], "mainnet", "test-agent").Return(handshake2, nil).Times(1)
	peerConnMng2.EXPECT().Start().Times(1)
	peerConnMng2.EXPECT().Sync().Times(1)
	peerConnMng2.EXPECT().GetPeerAddr().Return("127.0.0.1:6666").AnyTimes()
	peerConnMng2.EXPECT().StopSync().Times(1)
	peerConnMng2.EXPECT().Stop().Times(1)

	n, err := New("mainnet", "test-agent", newPeerConnMng, addrs, peerErrors, syncCompleted, handshakeManager, 10*time.Millisecond)
	require.NoError(t, err)

	n.Start()

	time.Sleep(10 * time.Millisecond) // what for all goroutines to start

	// when an error with the peer occurred, then the peerConManager is stop, sync process is stopped and reconnect again.
	// if we have only one peerConnection manager than the sync process will continue with it.
	peerErrors <- PeerErr{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}
	// receiving the same error second tie will not change the current peer in sync (127.0.0.2:6666)
	//peerErrors <- PeerErr{Peer: p2p.Peer{Address: "127.0.0.1:5555"}}

	time.Sleep(2 * time.Second) // wait all goroutines to restart again

	n.Stop()
}

func TestGetNextPeerConnMngForSync_HappyPath(t *testing.T) {
	pcm1 := &ServerPeer{network: "mainnet"}
	pcm2 := &ServerPeer{network: "testnet"}
	pcm3 := &ServerPeer{network: "fauced"}
	m := &sync.Map{}
	m.Store("127.0.0.1:3333", pcm1)
	m.Store("127.0.0.2:4444", pcm2)
	m.Store("127.0.0.3:5555", pcm3)

	n := Node{
		peerAddrs: []common.Addr{
			{IP: "127.0.0.1", Port: 3333},
			{IP: "127.0.0.2", Port: 4444},
			{IP: "127.0.0.3", Port: 5555},
		},
		peerChain:              m,
		getNextPeerConnMngWait: 10 * time.Millisecond,
	}
	actual, currentIndex := n.getNextPeerConnMngForSync(-1)

	require.Equal(t, 0, currentIndex)
	require.Equal(t, pcm1, actual)
}

func TestGetNextPeerConnMngForSync_WhenTheCurrentIndexIsEqualToLenOdSlice(t *testing.T) {
	pcm1 := &ServerPeer{network: "mainnet"}
	pcm2 := &ServerPeer{network: "testnet"}
	pcm3 := &ServerPeer{network: "fauced"}
	m := &sync.Map{}
	m.Store("127.0.0.1:3333", pcm1)
	m.Store("127.0.0.2:4444", pcm2)
	m.Store("127.0.0.3:5555", pcm3)

	n := Node{
		peerAddrs: []common.Addr{
			{IP: "127.0.0.1", Port: 3333},
			{IP: "127.0.0.2", Port: 4444},
			{IP: "127.0.0.3", Port: 5555},
		},
		peerChain:              m,
		getNextPeerConnMngWait: 10 * time.Millisecond,
	}
	actual, currentIndex := n.getNextPeerConnMngForSync(3)

	require.Equal(t, 0, currentIndex)
	require.Equal(t, pcm1, actual)
}

func TestGetNextPeerConnMngForSync_WhenThereIsNoAvailablePeerConnManagerAfterTheCurrentIndex(t *testing.T) {
	pcm1 := &ServerPeer{network: "mainnet"}
	pcm2 := &ServerPeer{network: "testnet"}

	m := &sync.Map{}
	m.Store("127.0.0.1:3333", pcm1)
	m.Store("127.0.0.2:4444", pcm2)

	n := Node{
		peerAddrs: []common.Addr{
			{IP: "127.0.0.1", Port: 3333},
			{IP: "127.0.0.2", Port: 4444},
			{IP: "127.0.0.3", Port: 5555},
		},
		peerChain:              m,
		getNextPeerConnMngWait: 10 * time.Millisecond,
	}
	actual, currentIndex := n.getNextPeerConnMngForSync(1)

	require.Equal(t, 0, currentIndex)
	require.Equal(t, pcm1, actual)
}

func TestGetNextPeerConnMngForSync_WhenThereIsNothingAvailableConnManagersThenTheMethodWillBlock(t *testing.T) {
	n := Node{
		peerAddrs: []common.Addr{
			{IP: "127.0.0.1", Port: 3333},
			{IP: "127.0.0.2", Port: 4444},
			{IP: "127.0.0.3", Port: 5555},
		},
		peerChain:              &sync.Map{},
		stop:                   make(chan struct{}),
		getNextPeerConnMngWait: 10 * time.Millisecond,
	}

	go func() {
		time.Sleep(20 * time.Millisecond) // give time the function to search for available conn managers
		close(n.stop)
	}()

	actual, currentIndex := n.getNextPeerConnMngForSync(0)

	require.Equal(t, 0, currentIndex)
	require.Nil(t, actual)
}

func TestSyncPeer_NotifySyncGoroutineForAnErrorInPeerThatIsNotInSync(t *testing.T) {
	ctrl := gomock.NewController(t)

	pcm1 := NewMockPeerConnectionManager(ctrl)
	pcm2 := NewMockPeerConnectionManager(ctrl)
	pcm1.EXPECT().Sync().Times(1)     // peer that is currently in sync process
	pcm1.EXPECT().StopSync().Times(1) // peer that is currently in sync process
	pcm1.EXPECT().GetPeerAddr().Return("127.0.0.2:3333").Times(1)
	pcm3 := NewMockPeerConnectionManager(ctrl)

	m := &sync.Map{}
	m.Store("127.0.0.1:3333", pcm1)
	m.Store("127.0.0.2:4444", pcm2)
	m.Store("127.0.0.3:5555", pcm3)

	n := Node{
		peerAddrs: []common.Addr{
			{IP: "127.0.0.1", Port: 3333},
			{IP: "127.0.0.2", Port: 4444},
			{IP: "127.0.0.3", Port: 5555},
		},
		peerChain:              m,
		stop:                   make(chan struct{}),
		doneSync:               make(chan struct{}),
		notifySyncForError:     make(chan PeerErr),
		getNextPeerConnMngWait: 10 * time.Millisecond,
	}
	go n.syncPeers()

	n.notifySyncForError <- PeerErr{
		Peer: p2p.Peer{
			Address: "127.0.0.3:5555", // peer that is not in sync process
		},
	}

	close(n.stop)
	<-n.doneSync
}

func TestSyncPeer_WhenThereIsNoAvailablePeerConnManagersUntilSomeTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	pcm := NewMockPeerConnectionManager(ctrl)
	pcm.EXPECT().Sync().Times(1)
	pcm.EXPECT().StopSync().Times(1)
	pcm.EXPECT().GetPeerAddr().Times(1)

	n := Node{
		peerAddrs: []common.Addr{
			{IP: "127.0.0.1", Port: 3333},
			{IP: "127.0.0.2", Port: 4444},
			{IP: "127.0.0.3", Port: 5555},
		},
		peerChain:              &sync.Map{},
		stop:                   make(chan struct{}),
		doneSync:               make(chan struct{}),
		notifySyncForError:     make(chan PeerErr),
		getNextPeerConnMngWait: 10 * time.Millisecond,
	}
	go n.syncPeers()

	time.Sleep(50 * time.Millisecond)        // give to search for available conn managers
	n.peerChain.Store("127.0.0.1:3333", pcm) // store conn manager

	n.notifySyncForError <- PeerErr{
		Peer: p2p.Peer{
			Address: "127.0.0.3:5555", // peer that is not in sync process
		},
	}

	close(n.stop)
	<-n.doneSync
}

func TestSyncPeer_WhenReceiveErrorFromPeerWhileSyncWithIt(t *testing.T) {
	ctrl := gomock.NewController(t)
	pcm1 := NewMockPeerConnectionManager(ctrl)
	pcm2 := NewMockPeerConnectionManager(ctrl)
	pcm1.EXPECT().Sync().Times(1) // currently in sync with it, but the node will receive error for this peer
	pcm1.EXPECT().StopSync().Times(1)
	pcm1.EXPECT().GetPeerAddr().Return("127.0.0.1:3333").Times(2) // this will be called when the node receive an error form the same peer that is in sync

	pcm2.EXPECT().Sync().Times(1) // when receive error this will be the next peer with each the node will continue to sync
	pcm2.EXPECT().StopSync().Times(1)
	pcm2.EXPECT().GetPeerAddr().Return("127.0.0.2:4444").Times(1) // this will be called when the node receive error from peer but not ffrom this one

	m := &sync.Map{}
	m.Store("127.0.0.1:3333", pcm1)
	m.Store("127.0.0.2:4444", pcm2)

	n := Node{
		peerAddrs: []common.Addr{
			{IP: "127.0.0.1", Port: 3333},
			{IP: "127.0.0.2", Port: 4444},
			{IP: "127.0.0.3", Port: 5555},
		},
		peerChain:              m,
		stop:                   make(chan struct{}),
		doneSync:               make(chan struct{}),
		notifySyncForError:     make(chan PeerErr),
		getNextPeerConnMngWait: 10 * time.Millisecond,
	}
	go n.syncPeers()

	pe := PeerErr{Peer: p2p.Peer{Address: "127.0.0.8:8888"}}
	n.notifySyncForError <- pe // first receive an error that is not with the peer which is in sync

	n.peerChain.Delete("127.0.0.1:3333") // delete the conn manager because an error is received from it.
	pe = PeerErr{Peer: p2p.Peer{Address: "127.0.0.1:3333"}}
	n.notifySyncForError <- pe // this peer is in sync and the Node will continue to the next one (127.0.0.2:4444)
	n.notifySyncForError <- pe // send the same error for the second time will not change the current peer (127.0.0.2:4444)

	close(n.stop)
	<-n.doneSync
}

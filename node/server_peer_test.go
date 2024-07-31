package node_test

import (
	"errors"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/node"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func TestServerPeer_StartHandleIncomingMsgsHeadersAndBlocks(t *testing.T) {
	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	bh1 := newBlockHeader(prevBlockHash)
	bh2 := newBlockHeader(node.Hash(bh1))
	bh3 := newBlockHeader(node.Hash(bh2))
	blockHeaders := []p2p.BlockHeader{bh1, bh2, bh3}
	msgHeaders := &p2p.MsgHeaders{Count: 3, BlockHeaders: blockHeaders}
	block := newMsgBlock(prevBlockHash)

	ctrl := gomock.NewController(t)

	fConn := &FakeConn{}
	msgHandlersManager := node.NewMockMsgHandlersManager(ctrl)
	msgHandlersManager.EXPECT().StartOverviewHandlers().Times(1)
	msgHandlersManager.EXPECT().Start().Times(1)
	msgHandlersManager.EXPECT().Stop().Times(1)
	peerSync := node.NewMockSyncManager(ctrl)
	peerSync.EXPECT().Start().Times(1)
	peerSync.EXPECT().Stop().Times(1)
	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
	networkMessageHandler.EXPECT().ReadMessage(fConn).Return(msgHeaders, nil).Times(1)
	networkMessageHandler.EXPECT().ReadMessage(fConn).Return(&block, nil).Times(1)

	peer := p2p.Peer{Connection: fConn, Address: "127.0.0.1:5555"}
	outgoingMsgs := make(chan *p2p.Message)
	errors := make(chan node.PeerErr)
	msgHeadersCh := make(chan *p2p.MsgHeaders)
	msgBlocksCh := make(chan *p2p.MsgBlock)

	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
	sp.Start()
	sp.Sync()

	actualHeaders := <-msgHeadersCh
	actualBlock := <-msgBlocksCh
	require.Equal(t, msgHeaders, actualHeaders)
	require.Equal(t, &block, actualBlock)

	sp.Stop()
	require.True(t, fConn.IsClosed)
}

func TestServerPeer_StartHandleOutgoingMsgsHeadersAndBlocks(t *testing.T) {
	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	msgGetHeaders, _ := p2p.NewMsgGetHeader("mainnet", 1, prevBlockHash, [32]byte{})
	b, _ := binary.Marshal(newMsgBlock(prevBlockHash))
	blockMsg, _ := p2p.NewMessage(p2p.CmdPong, "mainnet", b)

	ctrl := gomock.NewController(t)
	msgHandlersManager := node.NewMockMsgHandlersManager(ctrl)
	msgHandlersManager.EXPECT().StartOverviewHandlers().Times(1)
	msgHandlersManager.EXPECT().Stop().Times(1)
	peerSync := node.NewMockSyncManager(ctrl)
	peerSync.EXPECT().Stop()

	fConn := &FakeConn{}
	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
	networkMessageHandler.EXPECT().WriteMessage(msgGetHeaders, fConn).Return(nil).Times(1)
	networkMessageHandler.EXPECT().WriteMessage(blockMsg, fConn).Return(nil).Times(1)
	networkMessageHandler.EXPECT().ReadMessage(fConn).Return(&p2p.Message{}, &timeoutError{}).AnyTimes()

	peer := p2p.Peer{Connection: fConn, Address: "127.0.0.1:5555"}
	outgoingMsgs := make(chan *p2p.Message)
	errorsCh := make(chan node.PeerErr)
	msgHeadersCh := make(chan *p2p.MsgHeaders)
	msgBlocksCh := make(chan *p2p.MsgBlock)

	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
		networkMessageHandler, peer, outgoingMsgs, errorsCh, msgHeadersCh, msgBlocksCh)
	sp.Start()

	outgoingMsgs <- msgGetHeaders
	outgoingMsgs <- blockMsg

	require.Equal(t, 0, len(errorsCh))
	sp.Stop()
	require.True(t, fConn.IsClosed)
}

func TestServerPeer_WhenReadAndWriteTimeout(t *testing.T) {
	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	msgGetHeaders, _ := p2p.NewMsgGetHeader("mainnet", 1, prevBlockHash, [32]byte{})

	ctrl := gomock.NewController(t)
	msgHandlersManager := node.NewMockMsgHandlersManager(ctrl)
	msgHandlersManager.EXPECT().StartOverviewHandlers().Times(1)
	msgHandlersManager.EXPECT().Stop().Times(1)
	peerSync := node.NewMockSyncManager(ctrl)
	peerSync.EXPECT().Stop()

	fConn := &FakeConn{}
	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
	networkMessageHandler.EXPECT().WriteMessage(msgGetHeaders, fConn).Return(&timeoutError{}).Times(1)
	networkMessageHandler.EXPECT().ReadMessage(fConn).Return(&p2p.Message{}, &timeoutError{}).AnyTimes()

	peer := p2p.Peer{Connection: fConn, Address: "127.0.0.1:5555"}
	outgoingMsgs := make(chan *p2p.Message)
	errors := make(chan node.PeerErr)
	msgHeadersCh := make(chan *p2p.MsgHeaders)
	msgBlocksCh := make(chan *p2p.MsgBlock)

	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
	sp.Start()

	outgoingMsgs <- msgGetHeaders

	//time.Sleep(1 * time.Second)
	require.Equal(t, 0, len(errors))
	sp.Stop()
	require.True(t, fConn.IsClosed)
}

func TestServerPeer_WhenReadMsgFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	msgHandlersManager := node.NewMockMsgHandlersManager(ctrl)
	msgHandlersManager.EXPECT().StartOverviewHandlers().Times(1)
	msgHandlersManager.EXPECT().Stop().Times(1)
	peerSync := node.NewMockSyncManager(ctrl)
	peerSync.EXPECT().Stop()

	fConn := &FakeConn{}
	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
	networkMessageHandler.EXPECT().ReadMessage(fConn).Return(&p2p.Message{}, errors.New("err"))

	peer := p2p.Peer{Connection: fConn, Address: "127.0.0.1:5555"}
	outgoingMsgs := make(chan *p2p.Message)
	errorsCh := make(chan node.PeerErr)
	msgHeadersCh := make(chan *p2p.MsgHeaders)
	msgBlocksCh := make(chan *p2p.MsgBlock)

	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
		networkMessageHandler, peer, outgoingMsgs, errorsCh, msgHeadersCh, msgBlocksCh)
	sp.Start()

	actual := <-errorsCh
	require.Equal(t, 0, len(errorsCh))

	expected := node.PeerErr{Peer: peer, Err: fmt.Errorf("receive an error while reading from peer: err")}
	require.Equal(t, expected, actual)
	time.Sleep(10 * time.Millisecond)
	require.True(t, fConn.IsClosed)
}

func TestServerPeer_WhenWriteMsgFail(t *testing.T) {
	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	msgGetHeaders, _ := p2p.NewMsgGetHeader("mainnet", 1, prevBlockHash, [32]byte{})

	ctrl := gomock.NewController(t)
	msgHandlersManager := node.NewMockMsgHandlersManager(ctrl)
	msgHandlersManager.EXPECT().StartOverviewHandlers().Times(1)
	msgHandlersManager.EXPECT().Stop().Times(1)
	peerSync := node.NewMockSyncManager(ctrl)
	peerSync.EXPECT().Stop()

	fConn := &FakeConn{}
	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
	networkMessageHandler.EXPECT().ReadMessage(fConn).Return(&p2p.Message{}, &timeoutError{}).AnyTimes()
	networkMessageHandler.EXPECT().WriteMessage(msgGetHeaders, fConn).Return(errors.New("err"))

	peer := p2p.Peer{Connection: fConn, Address: "127.0.0.1:5555"}
	outgoingMsgs := make(chan *p2p.Message)
	errorsCh := make(chan node.PeerErr)
	msgHeadersCh := make(chan *p2p.MsgHeaders)
	msgBlocksCh := make(chan *p2p.MsgBlock)

	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
		networkMessageHandler, peer, outgoingMsgs, errorsCh, msgHeadersCh, msgBlocksCh)
	sp.Start()

	outgoingMsgs <- msgGetHeaders

	actual := <-errorsCh
	require.Equal(t, 0, len(errorsCh))

	expected := node.PeerErr{Peer: peer, Err: fmt.Errorf("receive an error while write to peer: err")}
	require.Equal(t, expected, actual)
	time.Sleep(10 * time.Millisecond)
	require.True(t, fConn.IsClosed)
}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

type FakeConn struct {
	IsClosed bool
}

func (c *FakeConn) Read(b []byte) (n int, err error)  { return 0, err }
func (c *FakeConn) Write(b []byte) (n int, err error) { return 0, err }
func (c *FakeConn) Close() error {
	c.IsClosed = true
	return nil
}
func (c *FakeConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *FakeConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *FakeConn) SetDeadline(t time.Time) error      { return nil }
func (c *FakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *FakeConn) SetWriteDeadline(t time.Time) error { return nil }

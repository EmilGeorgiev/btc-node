package node_test

import (
	"encoding/hex"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"testing"
)

func TestMmm(t *testing.T) {
	b, _ := hex.DecodeString("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")

	fmt.Printf("%x\n", p2p.Reverse(b))
}

//
//import (
//	"fmt"
//	errors2 "github.com/EmilGeorgiev/btc-node/errors"
//	"github.com/EmilGeorgiev/btc-node/network/p2p"
//	"github.com/EmilGeorgiev/btc-node/node"
//	"github.com/golang/mock/gomock"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//func TestServerPeer_StartHandleIncomingMsgsHeadersAndBlocks(t *testing.T) {
//	prevBlockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
//
//	bh1 := newBlockHeader(prevBlockHash)
//	bh2 := newBlockHeader(node.Hash(bh1))
//	bh3 := newBlockHeader(node.Hash(bh2))
//	blockHeaders := []p2p.BlockHeader{bh1, bh2, bh3}
//	msgHeaders := &p2p.MsgHeaders{Count: 3, BlockHeaders: blockHeaders}
//	block := newMsgBlock(prevBlockHash)
//
//	ctrl := gomock.NewController(t)
//
//	msgHandlersManager := node.NewMockStartStop(ctrl)
//	msgHandlersManager.EXPECT().Start().Times(1)
//	msgHandlersManager.EXPECT().Stop().Times(1)
//	peerSync := node.NewMockStartStop(ctrl)
//	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
//	networkMessageHandler.EXPECT().ReadMessage(nil).Return(msgHeaders, nil).Times(1)
//	networkMessageHandler.EXPECT().ReadMessage(nil).Return(&block, nil).Times(1)
//
//	peer := p2p.Peer{Address: "127.0.0.1:5555"}
//	outgoingMsgs := make(chan *p2p.Message)
//	errors := make(chan node.PeerErr)
//	msgHeadersCh := make(chan *p2p.MsgHeaders)
//	msgBlocksCh := make(chan *p2p.MsgBlock)
//
//	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
//		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
//	sp.Start()
//
//	actualHeaders := <-msgHeadersCh
//	actualBlock := <-msgBlocksCh
//	require.Equal(t, msgHeaders, actualHeaders)
//	require.Equal(t, &block, actualBlock)
//
//	sp.Stop()
//}
//
//func TestServerPeer_StartHandleOutgoingMsgsHeadersAndBlocks(t *testing.T) {
//	ping, _, _ := p2p.NewPingMsg("mainnet")
//
//	ctrl := gomock.NewController(t)
//	msgHandlersManager := node.NewMockStartStop(ctrl)
//	msgHandlersManager.EXPECT().Start().Times(1)
//	msgHandlersManager.EXPECT().Stop().Times(1)
//	peerSync := node.NewMockStartStop(ctrl)
//	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
//	networkMessageHandler.EXPECT().WriteMessage(ping, nil).Return(nil).Times(1)
//	networkMessageHandler.EXPECT().ReadMessage(nil).Return(&p2p.Message{}, &timeoutError{}).AnyTimes()
//
//	peer := p2p.Peer{Address: "127.0.0.1:5555"}
//	outgoingMsgs := make(chan *p2p.Message)
//	errors := make(chan node.PeerErr)
//	msgHeadersCh := make(chan *p2p.MsgHeaders)
//	msgBlocksCh := make(chan *p2p.MsgBlock)
//
//	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
//		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
//	sp.Start()
//
//	outgoingMsgs <- ping
//
//	require.Equal(t, 0, len(errors))
//	sp.Stop()
//}
//
//func TestServerPeer_WhenReadAndWritesTimeout(t *testing.T) {
//	ping, _, _ := p2p.NewPingMsg("mainnet")
//
//	ctrl := gomock.NewController(t)
//	msgHandlersManager := node.NewMockStartStop(ctrl)
//	msgHandlersManager.EXPECT().Start().Times(1)
//	msgHandlersManager.EXPECT().Stop().Times(1)
//	peerSync := node.NewMockStartStop(ctrl)
//	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
//	networkMessageHandler.EXPECT().WriteMessage(ping, nil).Return(&timeoutError{}).Times(1)
//	networkMessageHandler.EXPECT().ReadMessage(nil).Return(&p2p.Message{}, &timeoutError{}).AnyTimes()
//
//	peer := p2p.Peer{Address: "127.0.0.1:5555"}
//	outgoingMsgs := make(chan *p2p.Message)
//	errors := make(chan node.PeerErr)
//	msgHeadersCh := make(chan *p2p.MsgHeaders)
//	msgBlocksCh := make(chan *p2p.MsgBlock)
//
//	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
//		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
//	sp.Start()
//
//	outgoingMsgs <- ping
//
//	require.Equal(t, 0, len(errors))
//	sp.Stop()
//}
//
//func TestServerPeer_WhenReadMsgFail(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	msgHandlersManager := node.NewMockStartStop(ctrl)
//	msgHandlersManager.EXPECT().Start().Times(1)
//	msgHandlersManager.EXPECT().Stop().Times(1)
//	peerSync := node.NewMockStartStop(ctrl)
//	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
//	networkMessageHandler.EXPECT().ReadMessage(nil).Return(&p2p.MsgBlock{}, fmt.Errorf("err")).AnyTimes()
//
//	peer := p2p.Peer{Address: "127.0.0.1:5555"}
//	outgoingMsgs := make(chan *p2p.Message)
//	errors := make(chan node.PeerErr, 100)
//	msgHeadersCh := make(chan *p2p.MsgHeaders)
//	msgBlocksCh := make(chan *p2p.MsgBlock)
//
//	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
//		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
//	sp.Start()
//
//	actual := <-errors
//	expected := node.PeerErr{
//		Peer: peer,
//		Err:  errors2.NewE("receive an error while reading from peer: 127.0.0.1:5555.", fmt.Errorf("err"), true),
//	}
//	require.Equal(t, expected, actual)
//
//	sp.Stop()
//}
//
//func TestServerPeer_WhenWriteMsgFail(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	msgHandlersManager := node.NewMockStartStop(ctrl)
//	msgHandlersManager.EXPECT().Start().Times(1)
//	msgHandlersManager.EXPECT().Stop().Times(1)
//	peerSync := node.NewMockStartStop(ctrl)
//	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
//	networkMessageHandler.EXPECT().ReadMessage(nil).Return(&p2p.MsgBlock{}, &timeoutError{}).AnyTimes()
//	networkMessageHandler.EXPECT().WriteMessage(&p2p.Message{}, nil).Return(fmt.Errorf("err")).AnyTimes()
//
//	peer := p2p.Peer{Address: "127.0.0.1:5555"}
//	outgoingMsgs := make(chan *p2p.Message)
//	errors := make(chan node.PeerErr, 100)
//	msgHeadersCh := make(chan *p2p.MsgHeaders)
//	msgBlocksCh := make(chan *p2p.MsgBlock)
//
//	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
//		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
//	sp.Start()
//
//	outgoingMsgs <- &p2p.Message{}
//
//	actual := <-errors
//	expected := node.PeerErr{
//		Peer: peer,
//		Err:  errors2.NewE("receive an error while write to peer: 127.0.0.1:5555.", fmt.Errorf("err"), true),
//	}
//	require.Equal(t, expected, actual)
//
//	sp.Stop()
//}
//
//func TestServerPeer_Sync(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	msgHandlersManager := node.NewMockStartStop(ctrl)
//	msgHandlersManager.EXPECT().Start()
//	msgHandlersManager.EXPECT().Stop()
//	peerSync := node.NewMockStartStop(ctrl)
//	peerSync.EXPECT().Start().Times(1)
//	peerSync.EXPECT().Stop().Times(1)
//	networkMessageHandler := node.NewMockNetworkMessageHandler(ctrl)
//	networkMessageHandler.EXPECT().ReadMessage(nil).Return(&p2p.Message{}, &timeoutError{}).AnyTimes()
//
//	peer := p2p.Peer{Address: "127.0.0.1:5555"}
//	outgoingMsgs := make(chan *p2p.Message)
//	errors := make(chan node.PeerErr)
//	msgHeadersCh := make(chan *p2p.MsgHeaders)
//	msgBlocksCh := make(chan *p2p.MsgBlock)
//
//	sp := node.NewServerPeer("mainnet", msgHandlersManager, peerSync,
//		networkMessageHandler, peer, outgoingMsgs, errors, msgHeadersCh, msgBlocksCh)
//	sp.Start()
//	sp.Sync()
//
//	sp.StopSync()
//	sp.Stop()
//}
//
//func TestServerPeer_StartSyncWhenServerPeerIsNotStarted(t *testing.T) {
//	sp := node.NewServerPeer("mainnet", nil, nil,
//		nil, p2p.Peer{}, nil, nil, nil, nil)
//	sp.Sync()
//
//}
//
//func TestServerPeer_StartServerPeerTwoTimes(t *testing.T) {
//	ctrl := gomock.NewController(t)
//
//	msgHandlersManager := node.NewMockStartStop(ctrl)
//	msgHandlersManager.EXPECT().Start().Times(1)
//	msgHandlersManager.EXPECT().Stop().Times(1)
//
//	sp := node.NewServerPeer("mainnet", msgHandlersManager, nil,
//		nil, p2p.Peer{}, nil, nil, nil, nil)
//	sp.Start()
//	sp.Start()
//
//	sp.Stop()
//}
//
//type timeoutError struct{}
//
//func (e *timeoutError) Error() string   { return "i/o timeout" }
//func (e *timeoutError) Timeout() bool   { return true }
//func (e *timeoutError) Temporary() bool { return true }

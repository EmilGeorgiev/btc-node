package node_test

import (
	"testing"

	"github.com/EmilGeorgiev/btc-node/node"
	"github.com/golang/mock/gomock"
)

func TestNewMessageHandlersManager_Start(t *testing.T) {
	ctrl := gomock.NewController(t)

	msgBlockHeaders := node.NewMockStartStop(ctrl)
	msgBlockHeaders.EXPECT().Start().Times(1)
	msgHeadersHandler := node.NewMockStartStop(ctrl)
	msgHeadersHandler.EXPECT().Start().Times(1)

	manager := node.NewMessageHandlersManager([]node.StartStop{msgBlockHeaders, msgHeadersHandler}, nil)

	manager.Start()
}

func TestNewMessageHandlersManager_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)

	msgGetHeadersHandler := node.NewMockStartStop(ctrl)
	msgGetHeadersHandler.EXPECT().Stop().Times(1)
	msgGetBlockHandler := node.NewMockStartStop(ctrl)
	msgGetBlockHandler.EXPECT().Stop().Times(1)
	msgPingHandler := node.NewMockStartStop(ctrl)
	msgPingHandler.EXPECT().Stop().Times(1)

	manager := node.NewMessageHandlersManager([]node.StartStop{msgPingHandler, msgGetBlockHandler, msgGetHeadersHandler}, nil)

	manager.Stop()
}

func TestNewMessageHandlersManager_StartAndStopWithNoHandlers(t *testing.T) {
	manager := node.NewMessageHandlersManager([]node.StartStop{}, nil)
	manager.Start()
	manager.Stop()
}

func TestNewMessageHandlersManager_StartOverviewHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)

	msgBlockHeaders := node.NewMockStartStop(ctrl)
	msgBlockHeaders.EXPECT().Start().Times(0)
	msgHeadersHandler := node.NewMockStartStop(ctrl)
	msgHeadersHandler.EXPECT().Start().Times(1)

	manager := node.NewMessageHandlersManager([]node.StartStop{msgBlockHeaders, msgHeadersHandler}, []node.StartStop{msgHeadersHandler})

	manager.StartOverviewHandlers()
}

package node_test

import (
	"github.com/EmilGeorgiev/btc-node/node"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestNewMessageHandlersManager_Start(t *testing.T) {
	ctrl := gomock.NewController(t)

	msgGetHeadersHandler := node.NewMockStartStop(ctrl)
	msgGetHeadersHandler.EXPECT().Start().Times(1)
	msgGetBlockHandler := node.NewMockStartStop(ctrl)
	msgGetBlockHandler.EXPECT().Start().Times(1)
	msgPingHandler := node.NewMockStartStop(ctrl)
	msgPingHandler.EXPECT().Start().Times(1)

	manager := node.NewMessageHandlersManager([]node.StartStop{msgPingHandler, msgGetBlockHandler, msgGetHeadersHandler})

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

	manager := node.NewMessageHandlersManager([]node.StartStop{msgPingHandler, msgGetBlockHandler, msgGetHeadersHandler})

	manager.Stop()
}

func TestNewMessageHandlersManager_StartAndStopWithNoHandlers(t *testing.T) {
	manager := node.NewMessageHandlersManager([]node.StartStop{})
	manager.Start()
	manager.Stop()
}

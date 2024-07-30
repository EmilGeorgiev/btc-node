package node

import (
	"log"
)

type MessageHandlersManager struct {
	msgHandlers      []StartStop
	OverviewHandlers []StartStop
}

func NewMessageHandlersManager(handlers []StartStop, overviewHandlers []StartStop) *MessageHandlersManager {
	return &MessageHandlersManager{
		msgHandlers:      handlers,
		OverviewHandlers: overviewHandlers,
	}
}

func (m *MessageHandlersManager) Start() {
	for _, h := range m.msgHandlers {
		log.Printf("Start Managerhandlers' Handler of <type: %T\n", h)
		h.Start()
	}
}

func (m *MessageHandlersManager) Stop() {
	for _, h := range m.msgHandlers {
		log.Printf("Stop Managerhandlers' handler of type: %T\n", h)
		h.Stop()
	}
}

func (m *MessageHandlersManager) StartOverviewHandlers() {
	for _, h := range m.OverviewHandlers {
		log.Printf("Start Managerhandlers' Handler of <type>: %T\n", h)
		h.Start()
	}
}

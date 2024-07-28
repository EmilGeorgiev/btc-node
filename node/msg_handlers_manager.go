package node

import (
	"log"
	"sync/atomic"
)

type MessageHandlersManager struct {
	msgHandlers []StartStop
	isStarted   atomic.Bool
}

func NewMessageHandlersManager(handlers []StartStop) *MessageHandlersManager {
	return &MessageHandlersManager{
		msgHandlers: handlers,
	}
}

func (m *MessageHandlersManager) Start() {
	log.Println("Start Managerhandlers")
	if m.isStarted.Load() {
		return
	}
	log.Println("Start Managerhandlers 1111111111")
	m.isStarted.Store(true)
	for _, h := range m.msgHandlers {
		log.Printf("Start Managerhandlers' Handler of <type: %T\n", h)
		h.Start()
	}
}

func (m *MessageHandlersManager) Stop() {
	log.Println("Stop Managerhandlers")
	if !m.isStarted.Load() {
		return
	}
	log.Println("Stop Managerhandlers 1111111111")
	m.isStarted.Store(false)
	for _, h := range m.msgHandlers {
		log.Printf("Stop Managerhandlers' handler of type: %T\n", h)
		h.Stop()
	}
}

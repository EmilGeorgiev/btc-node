package node

import "sync/atomic"

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
	if m.isStarted.Load() {
		return
	}
	m.isStarted.Store(true)
	for _, h := range m.msgHandlers {
		h.Start()
	}
}

func (m *MessageHandlersManager) Stop() {
	if !m.isStarted.Load() {
		return
	}
	m.isStarted.Store(false)
	for _, h := range m.msgHandlers {
		h.Stop()
	}
}

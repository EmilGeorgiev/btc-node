package node

type StartStop interface {
	Start()
	Stop()
}

type MessageHandlersManager struct {
	msgHandlers []StartStop
}

func NewMessageHandlersManager(handlers []StartStop) MessageHandlersManager {
	return MessageHandlersManager{
		msgHandlers: handlers,
	}
}

func (m MessageHandlersManager) Start() {
	for _, h := range m.msgHandlers {
		h.Start()
	}
}

func (m MessageHandlersManager) Stop() {
	for _, h := range m.msgHandlers {
		h.Stop()
	}
}

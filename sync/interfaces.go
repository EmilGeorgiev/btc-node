package sync

import "github.com/EmilGeorgiev/btc-node/network/p2p"

type StartStop interface {
	Start()
	Stop()
}

type HeadersHandler interface {
	StartHandleMsgHeaders()
	Stop()
}

type BlockHandler interface {
	HandleBlockMessages()
}

type Node interface {
	GetPeerAddress() string
}

type HeaderRequester interface {
	RequestHeadersFromLastBlock() error
	RequestHeadersFromBlockHash([32]byte) error
}

type BlockRepository interface {
	Save(block p2p.MsgBlock) error
	Get(key [32]byte) (p2p.MsgBlock, error)
	GetLast() (p2p.MsgBlock, error)
}

type MsgSender interface {
	SendMsg(message p2p.Message, toPeer string) error
}

type BlockValidator interface {
	Validate(*p2p.MsgBlock) error
}

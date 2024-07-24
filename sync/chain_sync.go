package sync

import (
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type BlockRepository interface {
	Save(block p2p.MsgBlock) error
	Get(key [32]byte) (p2p.MsgBlock, error)
	GetLast() (p2p.MsgBlock, error)
}

type MsgSender interface {
	SendMsg(message p2p.Message) error
}

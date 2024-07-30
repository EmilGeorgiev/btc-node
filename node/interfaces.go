package node

import (
	"net"

	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type BlockRepository interface {
	Save(block p2p.MsgBlock) error
	Get(key [32]byte) (p2p.MsgBlock, error)
	GetLast() (p2p.MsgBlock, error)
}

type HandshakeManager interface {
	CreateOutgoingHandshake(addr common.Addr, network, userAgent string) (p2p.Handshake, error)
	CreateIncomingHandshake(network, userAgent string) (p2p.Handshake, error)
}

type Validator interface {
	Validate(block *p2p.MsgBlock) error
}

type StartStop interface {
	Start()
	Stop()
}

type MsgHandlersManager interface {
	StartStop
	StartOverviewHandlers()
}

type SyncManager interface {
	StartStop

	StartChainOverview(peerAddr string, cho chan common.ChainOverview)
}

// PeerConnectionManager defines the interface for managing peer connections in a Bitcoin network.
type PeerConnectionManager interface {
	StartStop

	// Sync initializes sync with the node to which the current implementation is connected.
	Sync()

	// StopSync ...
	StopSync()

	GetPeerAddr() string
	GetChainOverview() (<-chan common.ChainOverview, error)
}

type NetworkMessageHandler interface {
	ReadMessage(conn net.Conn) (interface{}, error)
	WriteMessage(msg *p2p.Message, conn net.Conn) error
}

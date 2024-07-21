package p2p

import (
	"fmt"
	"io"
)

type Peer struct {
	Address    string
	Connection io.ReadWriteCloser
	PongCh     chan uint64
	Services   uint64
	UserAgent  string
	Version    int32
}

// ID returns peer ID.
func (p Peer) ID() string {
	return p.Address
}

func (p Peer) String() string {
	return fmt.Sprintf("%s (%s)", p.UserAgent, p.Address)
}

type peerPing struct {
	nonce  uint64
	peerID string
}

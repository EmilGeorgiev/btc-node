package p2p

import (
	"math/rand"
	"time"
)

type MsgVersion struct {
	Version     int32
	Services    uint64
	Timestamp   int64
	AddrRecv    NetAddr
	AddrFrom    NetAddr
	Nonce       uint64
	UserAgent   VarStr
	StartHeight int32
	Relay       bool
}

// NewVersionMsg returns a new MsgVersion.
func NewVersionMsg(network, userAgent string, peerIP IPv4, peerPort uint16) (*Message, error) {
	payload := MsgVersion{
		Version:   Version,
		Services:  1,
		Timestamp: time.Now().UTC().Unix(),
		AddrRecv: NetAddr{
			Services: 1,
			IP:       peerIP,
			Port:     peerPort,
		},
		AddrFrom: NetAddr{
			Services: 1,
			IP:       *NewIPv4(127, 0, 0, 1),
			Port:     9333,
		},
		Nonce:       rand.Uint64(),
		UserAgent:   NewVarStr(userAgent),
		StartHeight: -1,
		Relay:       true,
	}

	msg, err := NewMessage("version", network, payload)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

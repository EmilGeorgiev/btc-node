package main

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/common"
	"math"
	"net"
	"time"
)

const (
	mainnet = "mainnet"
	simnet  = "simnet"
)

type Config struct {
	PeerAddrs              []common.Addr
	Network                string
	UserAgent              string
	DBPath                 string
	PingInterval           time.Duration
	PingTimeout            time.Duration
	ReadTimeout            time.Duration
	WriteTimeout           time.Duration
	SyncWait               time.Duration
	GetNextPeerConnMngWait time.Duration
}

func (c Config) Validate() error {
	for _, addr := range c.PeerAddrs {
		ip := net.ParseIP(addr.IP)
		if ip == nil {
			return fmt.Errorf("failed validating Config. The value: %s is not valid IP address", addr.IP)
		}

		if addr.Port < 0 || addr.Port > math.MaxUint16 {
			return fmt.Errorf("failed validating config. Port number: %d is not valid, it must be between 0 - %d", addr.Port, math.MaxUint16)
		}
	}

	if c.Network != mainnet && c.Network != simnet {
		return fmt.Errorf("failed validating config. Network: %s is not valid. Allowed values are [mainnet, simnet]", c.Network)
	}

	return nil
}

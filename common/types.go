package common

import (
	"fmt"
	"math/big"
)

type Addr struct {
	IP   string
	Port int64
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

type ChainOverview struct {
	Peer           string
	NumberOfBlocks int64
	CumulativeWork *big.Int
	IsValid        bool
}

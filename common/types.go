package common

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
)

var fileLog os.File

func WriteToFile(log string) {
	fileLog.WriteString(log)
}

type Addr struct {
	IP   string
	Port int64
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func AddrFromString(addr string) Addr {
	p := strings.Split(addr, ":")
	if len(p) != 2 {
		log.Printf("Invalid peer IP address: %s", addr)
		return Addr{}
	}
	port, err := strconv.Atoi(p[1])
	if err != nil {
		log.Printf("Invalid port number if IP address: %s", addr)
		return Addr{}
	}
	return Addr{
		IP:   p[0],
		Port: int64(port),
	}
}

type ChainOverview struct {
	Peer           string
	NumberOfBlocks int64
	CumulativeWork *big.Int
	IsValid        bool
}

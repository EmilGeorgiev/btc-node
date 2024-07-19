package network

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"io"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	version := p2p.MsgVersion{
		Version:   70015,
		Services:  1,
		Timestamp: time.Now().UTC().Unix(),
		AddrRecv: p2p.NetAddr{
			Services: 1,
			IP:       p2p.NewIPv4(185, 217, 241, 142),
			Port:     8333,
		},
		AddrFrom: p2p.NetAddr{
			Services: 1,
			IP:       p2p.NewIPv4(127, 0, 0, 1),
			Port:     8333, // dummy, we're not listening
		},
		Nonce:       rand.Uint64(),                               // returns a random number
		UserAgent:   p2p.NewVarStr("Satoshi:5.64/tinybit:0.0.1"), // returns a user-agent as a VarStr
		StartHeight: -1,
		Relay:       true,
	}

	fmt.Println("Creating a message")
	msg, err := p2p.NewMessage("version", "mainnet", version)
	if err != nil {
		panic(fmt.Sprintf("failed tocreate a message: %s\n", err))
	}

	fmt.Println("Serialize message")
	msgSerialized, err := msg.Serialize()
	if err != nil {
		panic(fmt.Sprintf("failed to serialie message: %s\n", err))
	}

	fmt.Println("open a tcp connection.")
	conn, err := net.Dial("tcp", "185.217.241.142:8333")
	defer conn.Close()

	fmt.Println("Sending MsgVersion to the node")
	if _, err = conn.Write(msgSerialized); err != nil {
		panic(fmt.Sprintf("failed to send version message to Node: %s\n", err))
	}

	tmp := make([]byte, 2048)

	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				panic(fmt.Sprintf("failed to read from connection: %s\n", err))
			}
			fmt.Printf("EOF stop the server: %s", err)
			return
		}
		fmt.Printf("received: %x\n", tmp[:n])
	}
}

package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/node"
	"testing"
)

// MsgVersion: {70015 1 1721451986 {0 1 0xc0000143b8 8333} {0 1 0xc0000143c0 8333} 6401031625427579582 {21 Satoshi:5.64/btc-node} -1 true}

func TestConnect(t *testing.T) {
	//version := p2p.MsgVersion{
	//	Version:   70015,
	//	Services:  1,
	//	Timestamp: time.Now().UTC().Unix(),
	//	AddrRecv: p2p.NetAddr{
	//		Services: 1,
	//		IP:       p2p.NewIPv4(185, 217, 241, 142),
	//		Port:     8333,
	//	},
	//	AddrFrom: p2p.NetAddr{
	//		Services: 1,
	//		IP:       p2p.NewIPv4(127, 0, 0, 1),
	//		Port:     8333, // dummy, we're not listening
	//	},
	//	Nonce:       rand.Uint64(),                               // returns a random number
	//	UserAgent:   p2p.NewVarStr("Satoshi:5.64/tinybit:0.0.1"), // returns a user-agent as a VarStr
	//	StartHeight: -1,
	//	Relay:       true,
	//}

	//fmt.Println("Creating a message")
	//msg, err := p2p.NewMessage("version", "mainnet", version)
	//if err != nil {
	//	panic(fmt.Sprintf("failed tocreate a message: %s\n", err))
	//}

	//fmt.Println("Serialize message")
	//msgSerialized, err := binary.Marshal(msg)
	//if err != nil {
	//	panic(fmt.Sprintf("failed to serialie message: %s\n", err))
	//}

	//fmt.Println("open a tcp connection.")
	//conn, err := net.Dial("tcp", "185.217.241.142:8333")
	//defer conn.Close()
	//
	//fmt.Println("Sending MsgVersion to the node")
	//if _, err = conn.Write(msgSerialized); err != nil {
	//	panic(fmt.Sprintf("failed to send version message to Node: %s\n", err))
	//}

	n, err := node.New("mainnet", "Satoshi:5.64/btc-node")
	if err != nil {
		panic(fmt.Sprintf("cann ot initialize node: %s", err.Error()))
	}

	fmt.Println("Running node")
	if err = n.Run(p2p.Addr{IP: "46.10.215.188", Port: 8333}); err != nil {
		panic(fmt.Sprintf("failed during running the node: %s", err))
	}
}

func TestBbb(t *testing.T) {
	var i uint64
	i = 1

	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, i)
	fmt.Println(err)

	fmt.Println(buf.Bytes())
}

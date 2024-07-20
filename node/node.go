package node

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"io"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type Node struct {
	Network      string
	NetworkMagic p2p.Magic
	UserAgent    string
}

// New returns a new Node.
func New(network, userAgent string) (*Node, error) {
	networkMagic, ok := p2p.Networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	return &Node{
		Network:      network,
		NetworkMagic: networkMagic,
		UserAgent:    userAgent,
	}, nil
}

// Run starts a node.
func (no Node) Run(nodeAddr string) error {
	peerAddr, err := ParseNodeAddr(nodeAddr)
	if err != nil {
		return fmt.Errorf("failed to parse node address: %s", err)
	}

	version := p2p.MsgVersion{
		Version:   p2p.Version,
		Services:  1,
		Timestamp: time.Now().UTC().Unix(),
		AddrRecv: p2p.NetAddr{
			Services: 1,
			IP:       peerAddr.IP,
			Port:     peerAddr.Port,
		},
		AddrFrom: p2p.NetAddr{
			Services: 1,
			IP:       *p2p.NewIPv4(127, 0, 0, 1),
			Port:     8333,
		},
		Nonce:       nonce(),
		UserAgent:   p2p.NewVarStr(no.UserAgent),
		StartHeight: -1,
		Relay:       true,
	}

	fmt.Println("Peerid: ", peerAddr.IP)
	fmt.Println("Peer port:", peerAddr.Port)

	fmt.Printf("MsgVersion: %v\n", version)

	msg, err := p2p.NewMessage("version", "mainnet", version)
	if err != nil {
		return fmt.Errorf("failed to create msgVersion: %s", err)
	}

	msgSerialized, err := binary.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal MsgVersion: %s", err)
	}

	fmt.Println("Open a connection to: ", nodeAddr)
	conn, err := net.Dial("tcp", nodeAddr)
	if err != nil {
		return fmt.Errorf("Failed to conenct to peer: %s ", err)
	}
	defer conn.Close()

	fmt.Println("Write MsgVersion to conenction")
	_, err = conn.Write(msgSerialized)
	if err != nil {
		return fmt.Errorf("Failed to write MsgVersion to conenction: %s ", err)
	}

	tmp := make([]byte, p2p.MsgHeaderLength)

Loop:
	for {
		fmt.Println("Read new message from connection")
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("faield to read from conenction: %s", err)
			}
			break Loop
		}

		fmt.Printf("received header: %x\n", tmp[:n])
		var msgHeader p2p.MessageHeader
		if err = binary.NewDecoder(bytes.NewReader(tmp[:n])).Decode(&msgHeader); err != nil {
			fmt.Printf("invalid header: %+v\n", err)
			continue
		}

		if err = msgHeader.Validate(); err != nil {
			fmt.Printf("Erro while validate message: %s\n", err.Error())
			continue
		}

		fmt.Printf("received message: %s\n", msgHeader.Command)

		switch msgHeader.CommandString() {
		case "version":
			if err = no.handleVersion(&msgHeader, conn); err != nil {
				fmt.Printf("failed to handle 'msgversion': %+v\n", err)
				continue
			}
		default:
			buf := make([]byte, msgHeader.Length)
			nn, err := conn.Read(buf)
			if err != nil {
				fmt.Printf("failed to read payalod: %s\n", err)
			}
			fmt.Printf("receive payalod: %x\n", buf[:nn])
		}
	}

	return nil
}

func nonce() uint64 {
	return rand.Uint64()
}

func (n Node) handleVersion(header *p2p.MessageHeader, conn io.ReadWriter) error {
	var version p2p.MsgVersion

	fmt.Println("Decode version message")
	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&version); err != nil {
		return err
	}

	fmt.Println("Create verack message")
	verack, err := p2p.NewVerackMsg(n.Network)
	if err != nil {
		return err
	}

	msg, err := binary.Marshal(verack)
	if err != nil {
		return err
	}

	fmt.Printf("Send verack message to peer")
	if _, err := conn.Write(msg); err != nil {
		return err
	}

	return nil
}

// Addr ...
type Addr struct {
	IP   p2p.IPv4
	Port uint16
}

// ParseNodeAddr ...
func ParseNodeAddr(nodeAddr string) (*Addr, error) {
	parts := strings.Split(nodeAddr, ":")
	if len(parts) != 2 {
		return nil, errors.New("malformed node address")
	}

	hostnamePart := parts[0]
	portPart := parts[1]
	if hostnamePart == "" || portPart == "" {
		return nil, errors.New("malformed node address")
	}

	port, err := strconv.Atoi(portPart)
	if err != nil {
		return nil, errors.New("malformed node address")
	}

	if port < 0 || port > math.MaxUint16 {
		return nil, errors.New("malformed node address")
	}

	var addr Addr
	ip := net.ParseIP(hostnamePart)
	copy(addr.IP[:], []byte(ip.To4()))

	addr.Port = uint16(port)

	return &addr, nil
}

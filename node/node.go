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

const (
	pingIntervalSec = 120
	pingTimeoutSec  = 30
)

type Node struct {
	Network      string
	NetworkMagic p2p.Magic
	UserAgent    string
	PingCh       chan peerPing
	PongCh       chan uint64
	Peers        map[string]*p2p.Peer
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
		Peers:        make(map[string]*p2p.Peer),
		PingCh:       make(chan peerPing),
		PongCh:       make(chan uint64),
	}, nil
}

// Run starts a node.
func (no Node) Run(nodeAddr string) error {
	peerAddr, err := ParseNodeAddr(nodeAddr)
	if err != nil {
		return fmt.Errorf("failed to parse node address: %s", err)
	}

	//version := p2p.MsgVersion{
	//	Version:   p2p.Version,
	//	Services:  1,
	//	Timestamp: time.Now().UTC().Unix(),
	//	AddrRecv: p2p.NetAddr{
	//		Services: 1,
	//		IP:       peerAddr.IP,
	//		Port:     peerAddr.Port,
	//	},
	//	AddrFrom: p2p.NetAddr{
	//		Services: 1,
	//		IP:       *p2p.NewIPv4(127, 0, 0, 1),
	//		Port:     8333,
	//	},
	//	Nonce:       nonce(),
	//	UserAgent:   p2p.NewVarStr(no.UserAgent),
	//	StartHeight: -1,
	//	Relay:       true,
	//}

	//fmt.Printf("MsgVersion: %v\n", version)

	version, err := p2p.NewVersionMsg(
		no.Network,
		no.UserAgent,
		peerAddr.IP,
		peerAddr.Port,
	)

	//msg, err := p2p.NewMessage("version", "mainnet", version)
	if err != nil {
		return fmt.Errorf("failed to create msgVersion: %s", err)
	}

	msgSerialized, err := binary.Marshal(version)
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

	go no.monitorPeers()

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
		case "verack":
			if err := no.handleVerack(&msgHeader, conn); err != nil {
				fmt.Printf("failed to handle 'verack': %+v\n", err)
				continue
			}
		case "ping":
			if err := no.handlePing(&msgHeader, conn); err != nil {
				fmt.Printf("failed to handle 'ping': %+v\n", err)
				continue
			}
		case "pong":
			if err := no.handlePong(&msgHeader, conn); err != nil {
				fmt.Printf("failed to handle 'pong': %+v\n", err)
				continue
			}
		case "inv":
			if err = no.handleInv(&msgHeader, conn); err != nil {
				fmt.Printf("failed to handle 'inv': %+v\n", err)
				continue
			}
		case "tx":
			if err = no.handleTx(&msgHeader, conn); err != nil {
				fmt.Printf("failed to handle 'tx': %+v\n", err)
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

func (n Node) handleVersion(header *p2p.MessageHeader, conn net.Conn) error {
	var version p2p.MsgVersion

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&version); err != nil {
		return err
	}

	peer := p2p.Peer{
		Address:    conn.RemoteAddr(),
		Connection: conn,
		PongCh:     make(chan uint64),
		Services:   version.Services,
		UserAgent:  version.UserAgent.String,
		Version:    version.Version,
	}
	n.Peers[peer.ID()] = &peer
	go n.monitorPeers()
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

func (n Node) handlePing(header *p2p.MessageHeader, conn io.ReadWriter) error {
	var ping p2p.MsgPing

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&ping); err != nil {
	}

	pong, err := p2p.NewPongMsg(n.Network, ping.Nonce)
	if err != nil {
		return err
	}

	msg, err := binary.Marshal(pong)
	if err != nil {
		return err
	}

	if _, err = conn.Write(msg); err != nil {
		return err
	}

	return nil
}

func (n Node) monitorPeers() {
	peerPings := make(map[uint64]string)

	for {
		select {
		case nonce := <-n.PongCh:
			peerID := peerPings[nonce]
			if peerID == "" {
				break
			}
			peer := n.Peers[peerID]
			if peer == nil {
				break
			}

			peer.PongCh <- nonce
			delete(peerPings, nonce)

		case pp := <-n.PingCh:
			peerPings[pp.nonce] = pp.peerID
		}
	}
}

func (n *Node) monitorPeer(peer *p2p.Peer) {
	for {
		time.Sleep(pingIntervalSec * time.Second)

		ping, nonce, err := p2p.NewPingMsg(n.Network)
		if err != nil {
			fmt.Printf("monitorPeer, NewPingMsg: %v\n", err)
		}

		msg, err := binary.Marshal(ping)
		if err != nil {
			fmt.Printf("monitorPeer, binary.Marshal: %v\n", err)
		}

		if _, err := peer.Connection.Write(msg); err != nil {
			n.disconnectPeer(peer.ID())
		}

		fmt.Printf("sent 'ping' to %s", peer)

		n.PingCh <- peerPing{
			nonce:  nonce,
			peerID: peer.ID(),
		}

		t := time.NewTimer(pingTimeoutSec * time.Second)

		select {
		case pn := <-peer.PongCh:
			if pn != nonce {
				fmt.Printf("nonce doesn't match for %s: want %d, got %d\n", peer, nonce, pn)
				n.disconnectPeer(peer.ID())
				return
			}
			fmt.Printf("got 'pong' from %s\n", peer)
		case <-t.C:
			// TODO: clean up peerPings, memory leak possible
			n.disconnectPeer(peer.ID())
			return
		}

		t.Stop()
	}
}

func (n Node) handlePong(header *p2p.MessageHeader, conn io.ReadWriter) error {
	var pong p2p.MsgPing

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&pong); err != nil {
		return err
	}

	n.PongCh <- pong.Nonce

	return nil
}

func (no Node) handleTx(header *p2p.MessageHeader, conn io.ReadWriter) error {
	var tx p2p.MsgTx

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&tx); err != nil {
		return err
	}

	fmt.Printf("transaction: %+v\n", tx)

	return nil
}

func (n Node) handleVerack(header *p2p.MessageHeader, conn io.ReadWriter) error {
	return nil
}

func (no Node) handleInv(header *p2p.MessageHeader, conn io.ReadWriter) error {
	var inv p2p.MsgInv

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&inv); err != nil {
		return err
	}

	var getData p2p.MsgGetData
	getData.Inventory = inv.Inventory
	getData.Count = inv.Count

	getDataMsg, err := p2p.NewMessage("getdata", no.Network, getData)
	if err != nil {
		return err
	}

	msg, err := binary.Marshal(getDataMsg)
	if err != nil {
		return err
	}

	_, err = conn.Write(msg)
	return err
}

func (no Node) disconnectPeer(peerID string) {
	fmt.Printf("disconnecting peer %s\n", peerID)

	peer := no.Peers[peerID]
	if peer == nil {
		return
	}

	peer.Connection.Close()
}

type peerPing struct {
	nonce  uint64
	peerID string
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

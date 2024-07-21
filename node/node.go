package node

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	pingIntervalSec = 120
	pingTimeoutSec  = 30
)

// 1. first conencto to the list of peers (hadnshake version/verack)
// 2. Get from the database to which block we have information locally.
// 3. when we have information what blocks we have locally we can send command/message "getblocks"
//         to get blocks from the begining or from some point in the chain and continue syncing our local
//         chain with other nodes' local chains of data.
// 4. receive inv message -
//

type Node struct {
	Network         string
	NetworkMagic    p2p.Magic
	UserAgent       string
	PingCh          chan peerPing
	PongCh          chan uint64
	Peers           map[string]*p2p.Peer
	stop            chan struct{}
	readConnTimeout time.Duration
	stopReports     chan StopReports
}

// New returns a new Node.
func New(network, userAgent string, readConnTimeout time.Duration) (*Node, error) {
	networkMagic, ok := p2p.Networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	return &Node{
		Network:         network,
		NetworkMagic:    networkMagic,
		UserAgent:       userAgent,
		Peers:           make(map[string]*p2p.Peer),
		readConnTimeout: readConnTimeout,
		stop:            make(chan struct{}),
		stopReports:     make(chan StopReports),
	}, nil
}

func (n Node) deletePeer(p p2p.Peer) {}
func (n Node) addPeer(peer p2p.Peer) {}

func (n Node) initilizeOutgoingConnection(peerAddr common.Addr) error {
	handshake, err := p2p.CreateHandshake(peerAddr, n.Network, n.UserAgent)
	if err != nil {
		log.Println("Errr: ", err.Error())
		return err
	}

	n.addPeer(handshake.Peer)
	go n.manageMessagesWithPeer(handshake.Peer)
	return nil
}

func (n Node) manageMessagesWithPeer(peer p2p.Peer) {
	defer func() {
		log.Println("Close the connection with peer: ", peer.Address)
		n.deletePeer(peer)
		peer.Connection.Close()
	}()

	go func() {
		conn := peer.Connection
		tmp := make([]byte, p2p.MsgHeaderLength)
		for {
			select {
			case <-n.stop:
				log.Println("Stop gracefully the goroutine that manage the connection with peer: ", peer.Address)
				return
			default:
				conn.SetReadDeadline(time.Now().Add(n.readConnTimeout))
				bn, err := conn.Read(tmp)
				if err != nil {
					var netErr net.Error
					if errors.As(err, &netErr) && netErr.Timeout() {
						continue
					}
					log.Println("receive an error while reading from connection with peer: ", peer.Address)
					return
				}
				if err = n.handleMessage(tmp[:bn], conn); err != nil {
					log.Println(err.Error())
					return
				}
			}
		}

	}()
}

func (n Node) handleMessage(headerRaw []byte, conn net.Conn) error {
	addr := conn.RemoteAddr().String()
	log.Printf("received msg header: %x\n", headerRaw)
	var msgHeader p2p.MessageHeader
	if err := binary.NewDecoder(bytes.NewReader(headerRaw)).Decode(&msgHeader); err != nil {
		log.Printf("failed to decode message header from peer: %s, due an error: %s\n", addr, err)
		return err
	}

	if err := msgHeader.Validate(); err != nil {
		log.Printf("receive invalid headet message from peer: %s. The header is invalid becasue: %s\n", addr, err)
		return err
	}

	fmt.Printf("received message: %s\n", msgHeader.Command)
	switch msgHeader.CommandString() {
	case "version":
		return fmt.Errorf("receive unexpected message Version from peer: %s that violate protocol. Close the connection", addr)
	case "verack":
		return fmt.Errorf("receive unexpected message Verackn from peer: %s that violate protocol. Close the connection", addr)
	case "ping":
		if err := n.handlePing(&msgHeader, conn); err != nil {
			return fmt.Errorf("failed to handle 'ping': %v\n", err)
		}
	case "pong":
		if err := n.handlePong(&msgHeader, conn); err != nil {
			fmt.Printf("failed to handle 'pong': %+v\n", err)
			return nil
		}
	case "inv":
		if err := n.handleInv(&msgHeader, conn); err != nil {
			return fmt.Errorf("failed to handle 'inv': %+v\n", err)
		}
	case "tx":
		if err := n.handleTx(&msgHeader, conn); err != nil {
			return fmt.Errorf("failed to handle 'tx': %+v\n", err)
		}
	default:
		log.Println("missing handler for message of type: ", msgHeader.CommandString())
		buf := make([]byte, msgHeader.Length)
		nn, err := conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read payalod. Err: %s\n", err)
		}
		fmt.Printf("the payalod of msg: %s is: %x\n", msgHeader.CommandString(), buf[:nn])
	}
	return nil
}

func (n Node) ConnectToPeers(peerAddrs []common.Addr) error {
	// initialize outgoing connections to the list of peers
	for _, peerAddr := range peerAddrs {
		n.initilizeOutgoingConnection(peerAddr)
	}
	return nil
}

func (n Node) Sync() {}
func (n Node) Stop() StopReports {
	close(n.stop)

	stopReports := <-n.stopReports
	return stopReports
}

type StopReports struct {
	LastBlockHash string
	Errors        []error
}

func nonce() uint64 {
	return rand.Uint64()
}

//func (n Node) handleVersion(header *p2p.MessageHeader, conn net.Conn) error {
//	var version p2p.MsgVersion
//
//	lr := io.LimitReader(conn, int64(header.Length))
//	if err := binary.NewDecoder(lr).Decode(&version); err != nil {
//		return err
//	}
//
//	peer := p2p.Peer{
//		Address:    conn.RemoteAddr(),
//		Connection: conn,
//		PongCh:     make(chan uint64),
//		Services:   version.Services,
//		UserAgent:  version.UserAgent.String,
//		Version:    version.Version,
//	}
//	n.Peers[peer.ID()] = &peer
//	go n.monitorPeers()
//	verack, err := p2p.NewVerackMsg(n.Network)
//	if err != nil {
//		return err
//	}
//
//	msg, err := binary.Marshal(verack)
//	if err != nil {
//		return err
//	}
//
//	fmt.Printf("Send verack message to peer")
//	if _, err := conn.Write(msg); err != nil {
//		return err
//	}
//
//	return nil
//}

func (n Node) handlePing(header *p2p.MessageHeader, conn net.Conn) error {
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

	log.Println("sending pong message to peer:", conn.RemoteAddr())
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

//// Addr ...
//type Addr struct {
//	IP   p2p.IPv4
//	Port uint16
//}
//
//// ParseNodeAddr ...
//func ParseNodeAddr(nodeAddr string) (*Addr, error) {
//	parts := strings.Split(nodeAddr, ":")
//	if len(parts) != 2 {
//		return nil, errors.New("malformed node address")
//	}
//
//	hostnamePart := parts[0]
//	portPart := parts[1]
//	if hostnamePart == "" || portPart == "" {
//		return nil, errors.New("malformed node address")
//	}
//
//	port, err := strconv.Atoi(portPart)
//	if err != nil {
//		return nil, errors.New("malformed node address")
//	}
//
//	if port < 0 || port > math.MaxUint16 {
//		return nil, errors.New("malformed node address")
//	}
//
//	var addr Addr
//	ip := net.ParseIP(hostnamePart)
//	copy(addr.IP[:], []byte(ip.To4()))
//
//	addr.Port = uint16(port)
//
//	return &addr, nil
//}

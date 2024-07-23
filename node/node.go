package node

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/common"
	errors2 "github.com/EmilGeorgiev/btc-node/errors"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var genesisBlockHash = [32]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x19, 0xd6, 0x68,
	0x9c, 0x08, 0x5a, 0xe1, 0x65, 0x83, 0x1e, 0x93,
	0x4f, 0xf7, 0x63, 0xae, 0x46, 0xa2, 0xa6, 0xc1,
	0x72, 0xb3, 0xf1, 0xb6, 0x0a, 0x8c, 0xe2, 0x6f,
}

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
	UserAgent       string
	Peers           map[string]*p2p.Peer
	stop            chan struct{}
	outgoingMsgs    chan p2p.Message
	readConnTimeout time.Duration
	errors          chan connErr
	stopReports     chan StopReports

	m sync.Mutex
}

// New returns a new Node.
func New(network, userAgent string, readConnTimeout time.Duration) (*Node, error) {
	_, ok := p2p.Networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	return &Node{
		Network:         network,
		UserAgent:       userAgent,
		Peers:           make(map[string]*p2p.Peer),
		readConnTimeout: readConnTimeout,
		stop:            make(chan struct{}),
		stopReports:     make(chan StopReports),
		errors:          make(chan connErr, 100),
		outgoingMsgs:    make(chan p2p.Message, 100),
	}, nil
}

func (n Node) ConnectToPeers(peerAddrs []common.Addr) {
	// initialize outgoing connections to the list of peers
	n.monitoringConnections()
	for _, peerAddr := range peerAddrs {
		n.initilizeOutgoingConnection(peerAddr)
	}
}

func (n Node) Sync() {

	msg, err := p2p.NewMsgGetHeader(n.Network, 1, genesisBlockHash, [32]byte{0})
	if err != nil {
		n.errors <- connErr{
			err: errors2.NewE("failed to crate getHeaders message", err),
		}
	}

	fmt.Println("SEND MSH HEADER to peer")
	n.outgoingMsgs <- *msg
}

func (n Node) Stop() {
	close(n.stop)

	run := true
	for run {
		n.m.Lock()
		if len(n.Peers) == 0 {
			run = false
		}
		n.m.Unlock()
	}
	log.Println("all goroutines are stopped")
}

type connErr struct {
	peer p2p.Peer
	err  error
}

func (n Node) monitoringConnections() {
	go func() {
		for {
			select {
			case <-n.stop:
				return
			case f := <-n.errors:
				// TODO: here can be implemented a logic for handling errors that rise during communication between peers.
				// For example we can implement a logic like:
				//      - reconnect to the peer depending of the error.
				//      - connect to another peer
				//      - save some data in DB
				//      - send notification
				//      - or something else
				//
				// For now we just log the error.
				var e errors2.E
				if errors.As(f.err, &e) {
					log.Println("Monitor print errE:", f.err.Error())
				} else {
					log.Println("Monitor print error: ", f)
				}

			}
		}
	}()
}

func (n Node) deletePeer(p p2p.Peer) {
	n.m.Lock()
	defer n.m.Unlock()
	delete(n.Peers, p.Address)

}

func (n Node) addPeer(peer p2p.Peer) {
	n.m.Lock()
	defer n.m.Unlock()
	n.Peers[peer.Address] = &peer
}

func (n Node) initilizeOutgoingConnection(peerAddr common.Addr) {
	handshake, err := p2p.CreateOutgoingHandshake(peerAddr, n.Network, n.UserAgent)
	if err != nil {
		fmt.Println("report error")
		n.errors <- connErr{
			peer: p2p.Peer{Address: peerAddr.String()},
			err:  err,
		}
		fmt.Println("after report error")
		return
	}
	fmt.Println("initialize connection with peer")
	n.addPeer(handshake.Peer)
	n.handleMessagesFromPeer(handshake.Peer)
	n.sendMessagesToPeer(handshake.Peer)
}

func (n Node) sendMessagesToPeer(peer p2p.Peer) {

	conn := peer.Connection
	go func() {
		defer func() {
			log.Println("Close the connection with peer: ", peer.Address)
			n.deletePeer(peer)
			peer.Connection.Close()
		}()

		for {
			select {
			case <-n.stop:
				log.Println("Stop goroutine that send messages from peer: ", peer.Address)
				return
			case msg := <-n.outgoingMsgs:
				rawMsg, err := binary.Marshal(msg)
				if err != nil {
					m := fmt.Sprintf("failed to marshal outgoing message: %s", msg.MessageHeader.CommandString())
					n.errors <- connErr{
						peer: peer,
						err:  errors2.NewE(m, err, true),
					}
					return
				}
				fmt.Printf("send message: %s, to peer: %s\n", msg.MessageHeader.CommandString(), peer.Address)
				conn.SetWriteDeadline(time.Now().Add(n.readConnTimeout))
				_, err = conn.Write(rawMsg)
				if err != nil {
					fmt.Println("faield to send msg: ", msg.MessageHeader.CommandString())
					m := fmt.Sprintf("receive an error while writing msg: %s to peer: %s.", msg.MessageHeader.CommandString(), peer.Address)
					n.errors <- connErr{
						peer: peer,
						err:  errors2.NewE(m, err, true),
					}
					return
				}
				fmt.Printf("send msg: %s successfully\n", msg.MessageHeader.CommandString())
			}
		}
	}()
}

func (n Node) handleMessagesFromPeer(peer p2p.Peer) {
	go func() {
		defer func() {
			log.Println("Close the connection with peer: ", peer.Address)
			n.deletePeer(peer)
			peer.Connection.Close()
		}()

		conn := peer.Connection
		tmp := make([]byte, p2p.MsgHeaderLength)
		for {
			select {
			case <-n.stop:
				log.Println("Stop goroutine that handle messages from peer: ", peer.Address)
				return
			default:
				conn.SetReadDeadline(time.Now().Add(n.readConnTimeout))
				bn, err := conn.Read(tmp)
				if err != nil {
					var netErr net.Error
					if errors.As(err, &netErr) && netErr.Timeout() {
						log.Println("timeout read")
						continue
					}
					n.errors <- connErr{
						peer: peer,
						err:  errors2.NewE(fmt.Sprintf("receive an error while reading from peer: %s.", peer.Address), err, true),
					}
					return
				}
				if err = n.handleMessage(tmp[:bn], conn); err != nil {
					n.errors <- connErr{peer: peer, err: err}
					return
				}
			}
		}
	}()
}

func (n Node) handleMessage(headerRaw []byte, conn net.Conn) error {
	addr := conn.RemoteAddr().String()
	log.Printf("received msg with header: %x\n", headerRaw)
	var msgHeader p2p.MessageHeader
	if err := binary.NewDecoder(bytes.NewReader(headerRaw)).Decode(&msgHeader); err != nil {
		return errors2.NewE(fmt.Sprintf("failed to decode message header from peer: %s.", addr), err)
	}

	fmt.Println("Message is of type: ", msgHeader.CommandString())

	if err := msgHeader.Validate(); err != nil {
		return errors2.NewE(fmt.Sprintf("receive invalid headet message from peer: %s.", addr), err)
	}

	fmt.Printf("received message: %s\n", msgHeader.Command)
	switch msgHeader.CommandString() {
	case "version":
		return errors2.NewE(fmt.Sprintf("receive unexpected msg Version from peer: %s that violate protocol.", addr))
	case "verack":
		return errors2.NewE(fmt.Sprintf("receive unexpected msg Verackn from peer: %s that violate protocol.", addr))
	case "ping":
		if err := n.handlePing(&msgHeader, conn); err != nil {
			return err
		}
	//case "pong":
	//	if err := n.handlePong(&msgHeader, conn); err != nil {
	//		fmt.Printf("failed to handle 'pong': %+v\n", err)
	//		return nil
	//	}
	case "inv":
		if err := n.handleInv(&msgHeader, conn); err != nil {
			return fmt.Errorf("failed to handle 'inv': %+v\n", err)
		}
	case "tx":
		if err := n.handleTx(&msgHeader, conn); err != nil {
			return fmt.Errorf("failed to handle 'tx': %+v\n", err)
		}
	case "headers":
		if err := n.handleHeaders(&msgHeader, conn); err != nil {
			return fmt.Errorf("failed to handle 'headers': %+v\n", err)
		}
	default:
		log.Println("missing handler for message of type: ", msgHeader.CommandString())
		buf := make([]byte, msgHeader.Length)
		nn, err := conn.Read(buf)
		if err != nil {
			return errors2.NewE(fmt.Sprintf("failed to read payalod of msg %s through connection.", msgHeader.CommandString()), err, true)
		}
		fmt.Printf("the payalod of msg: %s is: %x\n", msgHeader.CommandString(), buf[:nn])
	}
	return nil
}

func (n Node) handlePing(header *p2p.MessageHeader, conn net.Conn) error {
	var ping p2p.MsgPing

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&ping); err != nil {
		return errors2.NewE(fmt.Sprintf("failed to decode msg Ping from peer: %s", conn.RemoteAddr().String()), err)
	}

	pong, err := p2p.NewPongMsg(n.Network, ping.Nonce)
	if err != nil {
		return err
	}

	msg, err := binary.Marshal(pong)
	if err != nil {
		return errors2.NewE(fmt.Sprintf("failed to marshal pong msg for peer: %s", conn.RemoteAddr().String()), err)
	}

	log.Println("sending pong message to peer:", conn.RemoteAddr())
	if _, err = conn.Write(msg); err != nil {
		log.Println("failed to send pong")
		return errors2.NewE(fmt.Sprintf("failed sending pong msg through conn to peer: %s", conn.RemoteAddr().String()), err, true)
	}

	return nil
}

//func (n Node) handlePong(header *p2p.MessageHeader, conn io.ReadWriter) error {
//	var pong p2p.MsgPing
//
//	lr := io.LimitReader(conn, int64(header.Length))
//	if err := binary.NewDecoder(lr).Decode(&pong); err != nil {
//		return err
//	}
//
//	n.PongCh <- pong.Nonce
//
//	return nil
//}

func (n Node) handleTx(header *p2p.MessageHeader, conn io.ReadWriter) error {
	var tx p2p.MsgTx
	lr := io.LimitReader(conn, int64(header.Length))
	return binary.NewDecoder(lr).Decode(&tx)
}

func (n Node) handleVerack(header *p2p.MessageHeader, conn io.ReadWriter) error {
	return nil
}

func (n Node) handleInv(header *p2p.MessageHeader, conn io.ReadWriter) error {
	var inv p2p.MsgInv

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&inv); err != nil {
		return err
	}

	var getData p2p.MsgGetData
	getData.Inventory = inv.Inventory
	getData.Count = inv.Count

	getDataMsg, err := p2p.NewMessage("getdata", n.Network, getData)
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

func (n Node) handleHeaders(header *p2p.MessageHeader, conn io.ReadWriter) error {
	fmt.Println("HANDLE MSG HEADERS")
	var h p2p.MsgHeaders

	lr := io.LimitReader(conn, int64(header.Length))
	if err := binary.NewDecoder(lr).Decode(&h); err != nil {
		fmt.Println("FAILED DECODE MSG HEADERS: ", err.Error())
		return err
	}

	log.Println("handle msg Headers with block headers:")
	for i, bh := range h.BlockHeaders {
		log.Printf("Header: %d is: %+v\n", i, bh)
	}
	return nil
}

type StopReports struct {
	LastBlockHash string
	Errors        []error
}

//type peerPing struct {
//	nonce  uint64
//	peerID string
//}

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

//func (n Node) monitorPeers() {
//	peerPings := make(map[uint64]string)
//
//	for {
//		select {
//		case nonce := <-n.PongCh:
//			peerID := peerPings[nonce]
//			if peerID == "" {
//				break
//			}
//			peer := n.Peers[peerID]
//			if peer == nil {
//				break
//			}
//
//			peer.PongCh <- nonce
//			delete(peerPings, nonce)
//
//		case pp := <-n.PingCh:
//			peerPings[pp.nonce] = pp.peerID
//		}
//	}
//}

//func (n *Node) monitorPeer(peer *p2p.Peer) {
//	for {
//		time.Sleep(pingIntervalSec * time.Second)
//
//		ping, nonce, err := p2p.NewPingMsg(n.Network)
//		if err != nil {
//			fmt.Printf("monitorPeer, NewPingMsg: %v\n", err)
//		}
//
//		msg, err := binary.Marshal(ping)
//		if err != nil {
//			fmt.Printf("monitorPeer, binary.Marshal: %v\n", err)
//		}
//
//		if _, err := peer.Connection.Write(msg); err != nil {
//			n.disconnectPeer(peer.ID())
//		}
//
//		fmt.Printf("sent 'ping' to %s", peer)
//
//		n.PingCh <- peerPing{
//			nonce:  nonce,
//			peerID: peer.ID(),
//		}
//
//		t := time.NewTimer(pingTimeoutSec * time.Second)
//
//		select {
//		case pn := <-peer.PongCh:
//			if pn != nonce {
//				fmt.Printf("nonce doesn't match for %s: want %d, got %d\n", peer, nonce, pn)
//				n.disconnectPeer(peer.ID())
//				return
//			}
//			fmt.Printf("got 'pong' from %s\n", peer)
//		case <-t.C:
//			// TODO: clean up peerPings, memory leak possible
//			n.disconnectPeer(peer.ID())
//			return
//		}
//
//		t.Stop()
//	}
//}

//func (n Node) disconnectPeer(peerID string) {
//	fmt.Printf("disconnecting peer %s\n", peerID)
//
//	peer := n.Peers[peerID]
//	if peer == nil {
//		return
//	}
//
//	peer.Connection.Close()
//}

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

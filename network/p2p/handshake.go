package p2p

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/errors"
	"github.com/EmilGeorgiev/btc-node/network/binary"
)

// minimalSupportedVersion defines the minimal
// version of the Bitcoin protocol supported by this implementation.
const minimalSupportedVersion = 70016

// HandshakeManager manages the handshake process for incoming and outgoing connections.
type HandshakeManager struct {
}

// NewHandshakeManager creates a new instance of HandshakeManager.
func NewHandshakeManager() HandshakeManager {
	return HandshakeManager{}
}

// CreateIncomingHandshake handles the handshake process for incoming connections.
func (hi HandshakeManager) CreateIncomingHandshake(network, userAgent string) (Handshake, error) {
	return Handshake{}, nil
}

// CreateOutgoingHandshake initiates the handshake process with a remote peer.
func (hi HandshakeManager) CreateOutgoingHandshake(peerAddr common.Addr, network, userAgent string) (Handshake, error) {
	log.Println("Initialize handshake with peer: ", peerAddr.String())
	conn, err := net.Dial("tcp", peerAddr.String())
	if err != nil {
		msg := fmt.Sprintf("failed to connect to peer: %s", peerAddr.String())
		return Handshake{}, errors.NewE(msg, err, true)
	}

	msg, err := createMsgVersion(peerAddr, network, userAgent)
	if err != nil {
		return Handshake{}, err
	}

	log.Printf("Send MsgVersion to pear: %s: %x\n", peerAddr.String(), msg)
	if _, err = conn.Write(msg); err != nil {
		return Handshake{}, errors.NewE(fmt.Errorf("failed to send MsgVersion to the peer: %s ", peerAddr.String()), err, true)
	}

	msgHeader := make([]byte, MsgHeaderLength)
	versionMsgIsReceived := false
	wtxidrelayIsReceived := false
	var handshake Handshake
	for {
		n, err := conn.Read(msgHeader)
		if err != nil {
			return Handshake{}, errors.NewE(
				fmt.Sprintf("receive error while reading from connection from peer: %s.", peerAddr.String()), err, true,
			)
		}

		var header MessageHeader
		if err = binary.NewDecoder(bytes.NewReader(msgHeader[:n])).Decode(&header); err != nil {
			m := fmt.Sprintf("failed to decode header of received messages from peer: %s during the handshake", peerAddr.String())
			return Handshake{}, errors.NewE(m, err, true)
		}

		if err = header.Validate(); err != nil {
			m := fmt.Sprintf("Error while validate message header from peer: %s", peerAddr.String())
			return Handshake{}, errors.NewE(m, err, true)
		}

		switch header.CommandString() {
		case "version":
			log.Println("receive msg version from peer: ", peerAddr.String())
			if versionMsgIsReceived {
				log.Printf("message version is received more than once during the handhsake with peer: %s . The message will be ignored\n", peerAddr.String())
				continue
			}
			versionMsgIsReceived = true
			handshake, err = handleVersion(header, conn)
			if err != nil {
				return Handshake{}, err
			}

		case "wtxidrelay":
			wtxidrelayIsReceived = true
			log.Println("wtxidrelay is received")
		case "verack":
			log.Println("receive msg verack")
			if !wtxidrelayIsReceived {
				log.Println("received verack before wtxidrelay. verack will be discarded")
				continue
			}
			return handshake, nil
		default:
			log.Printf("receive unexpected message: %s. it will be ignored\n", header.CommandString())
		}
	}
}

// createMsgVersion creates a "version" message to initiate the handshake with a peer.
func handleVersion(msgHeader MessageHeader, conn net.Conn) (Handshake, error) {
	var version MsgVersion

	lr := io.LimitReader(conn, int64(msgHeader.Length))
	if err := binary.NewDecoder(lr).Decode(&version); err != nil {
		return Handshake{}, errors.NewE(
			fmt.Sprintf("failed decode MsgVersion from peer: %s.", conn.RemoteAddr().String()), err, true)
	}

	peer := Peer{
		Address:    conn.RemoteAddr().String(),
		Connection: conn,
		Services:   version.Services,
		UserAgent:  version.UserAgent.String,
		Version:    version.Version,
	}

	if minimalSupportedVersion > version.Version {
		return Handshake{}, errors.NewE(
			fmt.Sprintf("peer: %s support to old protocol version: %d. Minimum supported version is: %d",
				peer.Address, version.Version, minimalSupportedVersion))
	}

	// SEND wtxidrelay
	wtxidrelay, err := NewMessage("wtxidrelay", "mainnet", []byte{})
	if err != nil {
		fmt.Println("can not initilize wtxidrelay message")
		return Handshake{}, err
	}
	msg, err := binary.Marshal(wtxidrelay)
	if err != nil {
		return Handshake{}, errors.NewE(fmt.Sprintf("failed to marshal verack msg for peer %s", peer.Address), err)
	}
	fmt.Printf("Send wtxidrelay message to peer")
	if _, err := conn.Write(msg); err != nil {
		return Handshake{}, errors.NewE(
			fmt.Sprintf("failed to send verack message through conn to peer: %s", peer.Address), err, true)
	}

	// SEND verack
	verack, err := NewVerackMsg("mainnet")
	if err != nil {
		return Handshake{}, err
	}

	msg, err = binary.Marshal(verack)
	if err != nil {
		return Handshake{}, errors.NewE(fmt.Sprintf("failed to marshal verack msg for peer %s", peer.Address), err)
	}

	fmt.Printf("Send verack message to peer")
	if _, err := conn.Write(msg); err != nil {
		return Handshake{}, errors.NewE(
			fmt.Sprintf("failed to send verack message through conn to peer: %s", peer.Address), err, true)
	}

	return Handshake{Peer: peer}, nil
}

// Handshake represents the handshake state with a peer.
type Handshake struct {
	Peer Peer
}

func createMsgVersion(peerAddr common.Addr, network, userAgent string) ([]byte, error) {
	ip := net.ParseIP(peerAddr.IP)
	a := IPv4{}
	copy(a[:], ip.To4())
	version, err := NewVersionMsg(network, userAgent, a, uint16(peerAddr.Port))
	if err != nil {
		return nil, err
	}

	b, err := binary.Marshal(version)
	if err != nil {
		return nil, errors.NewE("failed to marshal MsgVersion.", err)
	}
	return b, nil
}

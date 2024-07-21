package p2p

import (
	"bytes"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"io"
	"log"
	"net"
)

func CreateHandshake(peerAddr Addr, network, userAgent string) (Handshake, error) {
	log.Println("Initialize handshake with peer: ", peerAddr.String())
	conn, err := net.Dial("tcp", peerAddr.String())
	if err != nil {
		return Handshake{}, fmt.Errorf("failed to connect to peer: %s with error: %s", peerAddr.String(), err)
	}

	msg, err := createMsgVersion(peerAddr, network, userAgent)
	if err != nil {
		return Handshake{}, err
	}

	log.Printf("Send MsgVersion to pear: %s: %x\n", peerAddr.String(), msg)
	if _, err = conn.Write(msg); err != nil {
		log.Println("00000:", err.Error())
		return Handshake{}, fmt.Errorf("failed to send MsgVersion to the peer: %s ", err)
	}

	fmt.Println("11111111")
	msgHeader := make([]byte, MsgHeaderLength)
	versionMsgIsReceived := false
	var handshake Handshake
	for {
		n, err := conn.Read(msgHeader)
		if err != nil {
			log.Printf("Receive error while reading from connection to peer: %s. Error: %s\n", peerAddr.String(), err)
			if err != io.EOF {
				return Handshake{}, fmt.Errorf("faield to read from connection: %s", err)
			}
			return Handshake{}, err
		}

		var header MessageHeader
		if err = binary.NewDecoder(bytes.NewReader(msgHeader[:n])).Decode(&header); err != nil {
			return Handshake{}, fmt.Errorf("failed to decode header of received messages from peer: %s during the handshake. Error: %s", peerAddr.String(), err)
		}

		if err = header.Validate(); err != nil {
			fmt.Printf("Error while validate message: %s\n", err.Error())
			return Handshake{}, fmt.Errorf("message header is not valid. Error")
		}

		switch header.CommandString() {
		case "version":
			log.Println("receive msg version from peer")
			if versionMsgIsReceived {
				log.Println("msg version is received for second time")
				return Handshake{}, fmt.Errorf("message version is received twice during the handhsake with peer: %s "+
					"which violate protocol", peerAddr.String())
			}
			versionMsgIsReceived = true
			handshake, err = handleVersion(header, conn)
			if err != nil {
				return Handshake{}, err
			}
		case "verack":
			log.Println("receive msg verack")
			if versionMsgIsReceived {
				return handshake, nil
			}
			log.Println("verack is received before msg version")
			return Handshake{}, fmt.Errorf("unexpected message of type verack is received as a first message during the handshake with peer: %s",
				peerAddr.String())
		default:
			log.Println("receive unexpected message:", header.CommandString())
			return Handshake{}, fmt.Errorf("unexpected message of type %s is received during the handshake with peer: %s. Only version and verack are accepted at this stage",
				header.CommandString(), peerAddr.String())
		}
	}
}

func handleVersion(msgHeader MessageHeader, conn net.Conn) (Handshake, error) {
	var version MsgVersion

	lr := io.LimitReader(conn, int64(msgHeader.Length))
	if err := binary.NewDecoder(lr).Decode(&version); err != nil {
		return Handshake{}, fmt.Errorf("failed decode MsgVersion from peer: %s. Error: %s", conn.RemoteAddr().String(), err)
	}

	peer := Peer{
		Address:    conn.RemoteAddr().String(),
		Connection: conn,
		Services:   version.Services,
		UserAgent:  version.UserAgent.String,
		Version:    version.Version,
	}

	log.Println("receive msg version. The node support protocol version:", version.Version)
	//for _, n := range Networks {
	//	if n == msgHeader.Magic {
	//
	//	}
	//}
	verack, err := NewVerackMsg("mainnet")
	if err != nil {
		return Handshake{}, err
	}

	msg, err := binary.Marshal(verack)
	if err != nil {
		return Handshake{}, err
	}

	fmt.Printf("Send verack message to peer")
	if _, err := conn.Write(msg); err != nil {
		return Handshake{}, err
	}

	return Handshake{Peer: peer}, nil
}

type Handshake struct {
	Peer Peer
}

// Addr ...
type Addr struct {
	IP   string
	Port int64
}

func (a Addr) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func createMsgVersion(peerAddr Addr, network, userAgent string) ([]byte, error) {
	ip := net.ParseIP(peerAddr.IP)
	a := IPv4{}
	copy(a[:], ip.To4())
	fmt.Printf("IPV4: %v\n", a)
	version, err := NewVersionMsg(network, userAgent, a, uint16(peerAddr.Port))
	if err != nil {
		return nil, err
	}

	b, err := binary.Marshal(version)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MsgVersion with an error: %s", err)
	}
	return b, nil
}

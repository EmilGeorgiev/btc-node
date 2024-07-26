package network

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MessageReadWriter struct {
	readConnTimeout  time.Duration
	writeConnTimeout time.Duration
}

func NewMessageReadWriter(rTimeout, wTimeout time.Duration) MessageReadWriter {
	return MessageReadWriter{
		readConnTimeout:  rTimeout,
		writeConnTimeout: wTimeout,
	}
}

func (ml MessageReadWriter) ReadMessage(conn net.Conn) (interface{}, error) {
	tmp := make([]byte, p2p.MsgHeaderLength)
	conn.SetReadDeadline(time.Now().Add(ml.readConnTimeout))
	for {
		bn, err := conn.Read(tmp)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				log.Println("timeout read")
				continue
			}
			return nil, err
		}
		return ml.handleMessage(tmp[:bn], conn)
	}
}

func (ml MessageReadWriter) handleMessage(headerRaw []byte, conn net.Conn) (interface{}, error) {
	log.Printf("received msg with header: %x\n", headerRaw)
	var msgHeader p2p.MessageHeader
	if err := binary.NewDecoder(bytes.NewReader(headerRaw)).Decode(&msgHeader); err != nil {
		return nil, err
	}

	payloadLength := int(msgHeader.Length)
	payload := make([]byte, 0, payloadLength)
	tmp := 1024
	if tmp > payloadLength {
		tmp = payloadLength
	}
	tempBuffer := make([]byte, tmp) // Temporary buffer for reading in chunks

	lr := io.LimitReader(conn, int64(payloadLength))

	for len(payload) < payloadLength {
		num, err := lr.Read(tempBuffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading the payload:", err)
			return nil, err
		}
		payload = append(payload, tempBuffer[:num]...)
	}
	if len(payload) != payloadLength {
		fmt.Printf("Expected to read %d bytes, but only read %d\n", payloadLength, len(payload))
		return nil, fmt.Errorf("Expected to read %d bytes, but only read %d\n", payloadLength, len(payload))
	}

	return ml.decodeMessage(payload, msgHeader.CommandString())
}

func (ml MessageReadWriter) decodeMessage(payload []byte, command string) (interface{}, error) {
	fmt.Printf("received message: %s\n", command)

	var msg interface{}
	switch command {
	case "version":
		msg = p2p.MsgVersion{}
	case "verack":
		msg = p2p.MsgVerack{}
	case "block":
		msg = p2p.MsgBlock{}
	case "ping":
		msg = p2p.MsgPing{}
	case "headers":
		msg = p2p.MsgHeaders{}
	default:
		log.Println("missing logic for message with command: ", command)
	}

	buf := bytes.NewBuffer(payload)
	if err := binary.NewDecoder(buf).Decode(&msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ml MessageReadWriter) WriteMessage(msg *p2p.Message, conn net.Conn) error {
	rawMsg, err := binary.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal outgoing message: %s", msg.MessageHeader.CommandString())
	}

	conn.SetWriteDeadline(time.Now().Add(ml.writeConnTimeout))
	for {
		if _, err = conn.Write(rawMsg); err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				log.Println("timeout read")
				continue
			}

			return fmt.Errorf("receive an error while sending msg: %s to peer", msg.MessageHeader.CommandString())
		}
		return nil
	}
}

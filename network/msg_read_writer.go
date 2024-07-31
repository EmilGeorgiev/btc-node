package network

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

// MessageReadWriter manages reading and writing messages over a network connection with specified timeouts.
type MessageReadWriter struct {
	readConnTimeout  time.Duration
	writeConnTimeout time.Duration
}

// NewMessageReadWriter creates a new MessageReadWriter with the given read and write timeouts.
func NewMessageReadWriter(rTimeout, wTimeout time.Duration) MessageReadWriter {
	return MessageReadWriter{
		readConnTimeout:  rTimeout,
		writeConnTimeout: wTimeout,
	}
}

// ReadMessage reads a message from the given network connection and decode it.
// The returned interface's type is one of MsgHeaders, MsgPing, MsgBlock and others
func (ml MessageReadWriter) ReadMessage(conn net.Conn) (interface{}, error) {
	tmp := make([]byte, p2p.MsgHeaderLength)
	conn.SetReadDeadline(time.Now().Add(ml.readConnTimeout))
	bn, err := conn.Read(tmp)
	if err != nil {
		return nil, err
	}
	return ml.handleMessage(tmp[:bn], conn)
}

// handleMessage processes the message header and reads the message payload from the connection.
func (ml MessageReadWriter) handleMessage(headerRaw []byte, conn net.Conn) (interface{}, error) {
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
		conn.SetReadDeadline(time.Now().Add(ml.readConnTimeout))
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
	buf := bytes.NewBuffer(payload)

	switch command {
	case "version":
		msg := p2p.MsgVersion{}
		if err := binary.NewDecoder(buf).Decode(&msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case "verack":
		msg := p2p.MsgVerack{}
		if err := binary.NewDecoder(buf).Decode(&msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case "block":
		msg := p2p.MsgBlock{}
		if err := binary.NewDecoder(buf).Decode(&msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case "ping":
		msg := p2p.MsgPing{}
		if err := binary.NewDecoder(buf).Decode(&msg); err != nil {
			return nil, err
		}
		return &msg, nil
	case "headers":
		msg := p2p.MsgHeaders{}
		if err := binary.NewDecoder(buf).Decode(&msg); err != nil {
			return nil, err
		}
		return &msg, nil
	default:
		log.Println("missing logic for message with command: ", command)
		return &p2p.Unknown{}, nil
	}
}

// WriteMessage writes a message to the given network connection.
func (ml MessageReadWriter) WriteMessage(msg *p2p.Message, conn net.Conn) error {
	rawMsg, err := binary.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal outgoing message: %s", msg.MessageHeader.CommandString())
	}

	conn.SetWriteDeadline(time.Now().Add(ml.writeConnTimeout))
	_, err = conn.Write(rawMsg)
	return err
}

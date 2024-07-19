package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	checksumLength = 4
	nodeNetwork    = 1
	magicLength    = 4
)

var (
	magicMainnet = [magicLength]byte{0xf9, 0xbe, 0xb4, 0xd9}
	magicSimnet  = [magicLength]byte{0x16, 0x1c, 0x14, 0x12}
	networks     = map[string][magicLength]byte{
		"mainnet": magicMainnet,
		"simnet":  magicSimnet,
	}
)

// MessagePayload ...
type MessagePayload interface {
	Serialize() ([]byte, error)
}

func NewMessage(cmd, network string, payload MessagePayload) (*Message, error) {
	serializedPayload, err := payload.Serialize()
	if err != nil {
		return nil, err
	}

	command, ok := commands[cmd]
	if !ok {
		return nil, fmt.Errorf("unsupported command %s", cmd)
	}

	magic, ok := networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	msg := Message{
		Magic:    magic,
		Command:  command,
		Length:   uint32(len(serializedPayload)),
		Checksum: checksum(serializedPayload),
		Payload:  serializedPayload,
	}

	return &msg, nil
}

type Message struct {
	Magic    [magicLength]byte
	Command  [commandLength]byte
	Length   uint32
	Checksum [checksumLength]byte
	Payload  []byte
}

func (msg Message) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	if _, err := buf.Write(msg.Magic[:]); err != nil {
		return nil, err
	}

	if _, err := buf.Write(msg.Command[:]); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.LittleEndian, msg.Length); err != nil {
		return nil, err
	}

	if _, err := buf.Write(msg.Checksum[:]); err != nil {
		return nil, err
	}

	if _, err := buf.Write(msg.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

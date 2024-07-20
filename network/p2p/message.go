package p2p

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"strings"
)

const (
	checksumLength = 4
	nodeNetwork    = 1
	magicLength    = 4

	// MsgHeaderLength specifies the length of Message in bytes
	MsgHeaderLength = magicLength + commandLength + checksumLength + 4 // 4 - payload length value
)

var (
	MagicMainnet Magic = [magicLength]byte{0xf9, 0xbe, 0xb4, 0xd9}
	MagicSimnet  Magic = [magicLength]byte{0x16, 0x1c, 0x14, 0x12}
	Networks           = map[string][magicLength]byte{
		"mainnet": MagicMainnet,
		"simnet":  MagicSimnet,
	}
)

type Magic [magicLength]byte

var (
	magicMainnet = [magicLength]byte{0xf9, 0xbe, 0xb4, 0xd9}
	magicSimnet  = [magicLength]byte{0x16, 0x1c, 0x14, 0x12}
	networks     = map[string][magicLength]byte{
		"mainnet": magicMainnet,
		"simnet":  magicSimnet,
	}
)

type MessageHeader struct {
	Magic    [magicLength]byte
	Command  [commandLength]byte
	Length   uint32
	Checksum [checksumLength]byte
}

func (mh MessageHeader) Validate() error {
	if !mh.HasValidMagic() {
		return fmt.Errorf("invalid magic: %x", mh.Magic)
	}

	if !mh.HasValidCommand() {
		return fmt.Errorf("invalid command: %+v", mh.CommandString())
	}

	return nil
}

func (mh MessageHeader) HasValidCommand() bool {
	_, ok := commands[mh.CommandString()]
	return ok
}

func (mh MessageHeader) HasValidMagic() bool {
	switch mh.Magic {
	case MagicMainnet, MagicSimnet:
		return true
	}

	return false
}

func (mh MessageHeader) CommandString() string {
	return strings.Trim(string(mh.Command[:]), string(0))
}

// MessagePayload ...
type MessagePayload interface {
	Serialize() ([]byte, error)
}

func NewMessage(cmd, network string, payload interface{}) (*Message, error) {
	serializedPayload, err := binary.Marshal(payload)
	if err != nil {
		return nil, err
	}

	command, ok := commands[cmd]
	if !ok {
		return nil, fmt.Errorf("unsupported command %s", cmd)
	}

	magic, ok := Networks[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network %s", network)
	}

	msg := Message{
		MessageHeader: MessageHeader{
			Magic:    magic,
			Command:  command,
			Length:   uint32(len(serializedPayload)),
			Checksum: checksum(serializedPayload),
		},
		Payload: serializedPayload,
	}

	return &msg, nil
}

type Message struct {
	MessageHeader
	Payload []byte
}

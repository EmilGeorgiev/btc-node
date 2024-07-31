package p2p

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/errors"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"strings"
)

// Constants defining the length of various components in the Bitcoin protocol.
const (
	checksumLength = 4
	magicLength    = 4

	// MsgHeaderLength specifies the length of Message in bytes.
	// magicLength + CommandLength + checksumLength + 4 (payload length value)
	MsgHeaderLength = 24
)

// Predefined magic values for mainnet and simnet.
var (
	MagicMainnet Magic = [magicLength]byte{0xf9, 0xbe, 0xb4, 0xd9}
	MagicSimnet  Magic = [magicLength]byte{0x16, 0x1c, 0x14, 0x12}
	Networks           = map[string][magicLength]byte{
		"mainnet": MagicMainnet,
		"simnet":  MagicSimnet,
	}
)

// Magic represents the magic value in a Bitcoin message header.
type Magic [magicLength]byte

// Redundant variable definitions.
var (
	magicMainnet = [magicLength]byte{0xf9, 0xbe, 0xb4, 0xd9}
	magicSimnet  = [magicLength]byte{0x16, 0x1c, 0x14, 0x12}
	networks     = map[string][magicLength]byte{
		"mainnet": magicMainnet,
		"simnet":  magicSimnet,
	}
)

// MessageHeader represents the header of a Bitcoin protocol message.
type MessageHeader struct {
	// Magic value indicating message origin network,
	// and used to seek to next message when stream state is unknown
	Magic [magicLength]byte

	// ASCII string identifying the packet content,
	// NULL padded (non-NULL padding results in packet rejected)
	Command [CommandLength]byte

	// Length of payload in number of bytes
	Length uint32

	// First 4 bytes of sha256(sha256(payload))
	Checksum [checksumLength]byte
}

// Validate checks if the message header has a valid magic value.
func (mh MessageHeader) Validate() error {
	if !mh.HasValidMagic() {
		return fmt.Errorf("invalid magic: %x", mh.Magic)
	}

	return nil
}

// HasValidCommand checks if the message header has a valid command.
func (mh MessageHeader) HasValidCommand() bool {
	_, ok := commands[mh.CommandString()]
	return ok
}

// HasValidMagic checks if the magic value in the message header is valid.
func (mh MessageHeader) HasValidMagic() bool {
	switch mh.Magic {
	case MagicMainnet, MagicSimnet:
		return true
	}

	return false
}

// CommandString returns the command as a string, trimmed of any null characters.
func (mh MessageHeader) CommandString() string {
	return strings.TrimRight(string(mh.Command[:]), "\x00")
	//return strings.Trim(string(mh.Command[:]), string(0))
}

// MessagePayload ...
type MessagePayload interface {
	Serialize() ([]byte, error)
}

// MessagePayload represents the payload of a Bitcoin protocol message.
func NewMessage(cmd, network string, payload interface{}) (*Message, error) {
	serializedPayload, err := binary.Marshal(payload)
	if err != nil {
		msg := fmt.Sprintf("failed to create new message of type %s when parse its paylaod.", cmd)
		return nil, errors.NewE(msg, err)
	}

	command, ok := commands[cmd]
	if !ok {
		msg := fmt.Sprintf("failed to create a new message because the command is unsuported %s", cmd)
		return nil, errors.NewE(msg, err)
	}

	magic, ok := Networks[network]
	if !ok {
		msg := fmt.Sprintf("faile creating message if type %s becasue network: %s is not supported", cmd, network)
		return nil, errors.NewE(msg, err)
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

// Message represents a Bitcoin protocol message, including its header and payload.
type Message struct {
	MessageHeader
	Payload []byte
}

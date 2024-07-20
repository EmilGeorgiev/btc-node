package binary_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/EmilGeorgiev/btc-node/network/binary"
)

type customType []byte

func (ct customType) MarshalBinary() ([]byte, error) {
	return []byte{0xde, 0xad, 0xbe, 0xef}, nil
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		err      error
		expected []byte
	}{
		{name: "uint8",
			input:    uint8(255),
			err:      nil,
			expected: []byte{0xFF}},

		{name: "int32",
			input:    int32(1337),
			err:      nil,
			expected: []byte{0x39, 0x05, 0, 0}},

		{name: "uint32",
			input:    uint32(1337),
			err:      nil,
			expected: []byte{0x39, 0x05, 0, 0}},

		{name: "int64",
			input:    int64(1337),
			err:      nil,
			expected: []byte{0x39, 0x05, 0, 0, 0, 0, 0, 0}},

		{name: "uint64",
			input:    int64(1337),
			err:      nil,
			expected: []byte{0x39, 0x05, 0, 0, 0, 0, 0, 0}},

		{name: "bool true",
			input:    true,
			err:      nil,
			expected: []byte{0x01}},

		{name: "bool false",
			input:    false,
			err:      nil,
			expected: []byte{0}},

		{name: "magic or checksum",
			input:    [4]byte{0x31, 0x33, 0x70, 0},
			err:      nil,
			expected: []byte{0x31, 0x33, 0x70, 0}},

		{name: "command",
			input:    [12]byte{0x74, 0x65, 0x73, 0x74, 0, 0, 0, 0, 0, 0, 0, 0},
			err:      nil,
			expected: []byte{0x74, 0x65, 0x73, 0x74, 0, 0, 0, 0, 0, 0, 0, 0}},

		{name: "byte array",
			input:    []byte{0x12, 0x34, 0x56, 0x78},
			err:      nil,
			expected: []byte{0x12, 0x34, 0x56, 0x78}},

		{name: "struct",
			input: struct {
				Test  uint32
				Magic [4]byte
				Data  []byte
			}{
				Test:  31337,
				Magic: [4]byte{0x12, 0x34, 0x56, 0x78},
				Data:  []byte{0xde, 0xad, 0xbe, 0xef}},
			err:      nil,
			expected: []byte{0x69, 0x7A, 0, 0, 0x12, 0x34, 0x56, 0x78, 0xde, 0xad, 0xbe, 0xef}},

		{name: "struct with a pointer",
			input: struct {
				Test    uint32
				Pointer *struct {
					Test uint32
				}
			}{
				Test: 31337,
				Pointer: &struct {
					Test uint32
				}{
					Test: 31337,
				}},
			err:      nil,
			expected: []byte{0x69, 0x7A, 0, 0, 0x69, 0x7A, 0, 0}},

		{name: "struct with a string",
			input: struct {
				Test   uint32
				String string
			}{
				Test:   31337,
				String: "test"},
			err:      nil,
			expected: []byte{0x69, 0x7a, 0, 0, 0x74, 0x65, 0x73, 0x74}},

		{name: "custom marshaler",
			input:    customType{},
			err:      nil,
			expected: []byte{0xde, 0xad, 0xbe, 0xef}},

		{name: "unsupported type",
			input:    [3]byte{0x12, 0x34, 0x56},
			err:      errors.New("unsupported type [3]uint8"),
			expected: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			actual, err := binary.Marshal(test.input)
			if err == nil && test.err != nil {
				tt.Errorf("expected error: %v, actual: %s", test.err, actual)
				return
			}

			if err != nil && test.err == nil {
				tt.Errorf("didn't expect an error: %v", err)
				return
			}

			if !bytes.Equal(actual, test.expected) {
				tt.Errorf("expected: %x, actual %x", test.expected, actual)
			}
		})
	}
}

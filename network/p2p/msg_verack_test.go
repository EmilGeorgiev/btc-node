package p2p_test

import (
	"errors"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"reflect"
	"testing"
)

func TestNewVerackMsg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		err      error
		expected *p2p.Message
	}{
		{name: "ok",
			input: "mainnet",
			err:   nil,
			expected: &p2p.Message{
				MessageHeader: p2p.MessageHeader{
					Magic:    [4]byte{249, 190, 180, 217},
					Command:  [12]byte{118, 101, 114, 97, 99, 107, 0, 0, 0, 0, 0, 0},
					Length:   uint32(0),
					Checksum: [4]byte{93, 246, 224, 226},
				},
				Payload: []byte{},
			}},
		{name: "unsupported network",
			input:    "unknown",
			err:      errors.New("unsupported network 'unknown'"),
			expected: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			actual, err := p2p.NewVerackMsg(test.input)
			if err != nil && test.err == nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if err == nil && test.err != nil {
				t.Errorf("expected error: %+v, got: %+v", err, actual)
			}
			if err != nil && test.err != nil && err.Error() != test.err.Error() {
				t.Errorf("expected error: %+v, got: %+v", err, test.err)
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected: %+v, actual: %+v", test.expected, actual)
			}
		})
	}
}

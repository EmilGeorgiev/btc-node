package p2p

import (
	"bytes"
	"testing"
)

func TestVarInt_MarshalBinary(t *testing.T) {
	tests := []struct {
		name string
		v    VarInt
		want []byte
	}{
		{"Value less than 0xFD", VarInt(0xFC), []byte{0xFC}},
		{"Value equal to 0xFD", VarInt(0xFD), []byte{0xFD, 0xFD, 0x00}},
		{"Value equal to 0xFE", VarInt(0xFE), []byte{0xFD, 0xFE, 0x00}},
		{"Value equal to 0xFFFF", VarInt(0xFFFF), []byte{0xFD, 0xFF, 0xFF}},
		{"Value greater than 0xFFFF and less then <= 0xFFFFFFFF", VarInt(0xFFFFFFFC), []byte{0xFE, 0xFC, 0xFF, 0xFF, 0xFF}},
		{"Value equal to 0xFFFFFFFF", VarInt(0xFFFFFFFC), []byte{0xFE, 0xFC, 0xFF, 0xFF, 0xFF}},
		{"Value greater then 0xFFFFFFFF", VarInt(0xFFFFFFFFFC), []byte{0xFF, 0xFC, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.v.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary() error = %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("MarshalBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVarInt_UnmarshalBinary(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    VarInt
		wantErr bool
	}{
		{"Value less than 0xFD", []byte{0xFC}, VarInt(0xFC), false},
		{"Value equal to 0xFD", []byte{0xFD, 0xFD, 0x00}, VarInt(0xFD), false},
		{"Value equal to 0xFE", []byte{0xFD, 0xFE, 0x00}, VarInt(0xFE), false},
		{"Value greater than 0xFE", []byte{0xFE, 0xFF, 0xFF, 0x00, 0x00}, VarInt(0xFFFF), false},
		{"Large value 0xFFFFFFFF", []byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF}, VarInt(0xFFFFFFFF), false},
		{"Larger value 0xFFFFFFFFFF", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00}, VarInt(0xFFFFFFFF), false},
		{"Incomplete data for 0xFD", []byte{0xFD, 0x01}, 0, true},
		{"Incomplete data for 0xFE", []byte{0xFE, 0x01, 0x02, 0x03}, 0, true},
		{"Incomplete data for 0xFF", []byte{0xFF, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v VarInt
			err := v.UnmarshalBinary(bytes.NewReader(tt.data))
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if v != tt.want {
				t.Errorf("UnmarshalBinary() = %v, want %v", v, tt.want)
			}
		})
	}
}

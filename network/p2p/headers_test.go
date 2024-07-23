package p2p

import (
	"bytes"
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMsgHeaders_UnmarshalBinary(t *testing.T) {
	var rowMsg = "0102000000b6ff0b1b1680a2862a30ca44d346d9e8910d334beb48ca0c00000000000000009d10aa52ee949386ca9385695f04ede270dda20810decd12bc9b048aaab3147124d95a5430c31b18fe9f086400"
	b, _ := hex.DecodeString(rowMsg)
	mh := &MsgHeaders{}

	err := mh.UnmarshalBinary(bytes.NewReader(b))
	require.NoError(t, err)
	want := &MsgHeaders{
		Count: 1,
		BlockHeaders: []BlockHeader{
			{
				Version:       2,
				PrevBlockHash: [32]byte{0xb6, 0xff, 0x0b, 0x1b, 0x16, 0x80, 0xa2, 0x86, 0x2a, 0x30, 0xca, 0x44, 0xd3, 0x46, 0xd9, 0xe8, 0x91, 0x0d, 0x33, 0x4b, 0xeb, 0x48, 0xca, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				MerkleRoot:    [32]byte{0x9d, 0x10, 0xaa, 0x52, 0xee, 0x94, 0x93, 0x86, 0xca, 0x93, 0x85, 0x69, 0x5f, 0x04, 0xed, 0xe2, 0x70, 0xdd, 0xa2, 0x08, 0x10, 0xde, 0xcd, 0x12, 0xbc, 0x9b, 0x04, 0x8a, 0xaa, 0xb3, 0x14, 0x71},
				Timestamp:     1415239972,
				Bits:          404472624,
				Nonce:         1678286846,
				TxnCount:      0,
			},
		},
	}
	require.Equal(t, want, mh)
}

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

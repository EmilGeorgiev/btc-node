package p2p

import (
	"bytes"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"io"
)

type VarInt uint64

func (v *VarInt) UnmarshalBinary(r io.Reader) error {
	d := binary.NewDecoder(r)
	var i uint8
	if err := d.Decode(&i); err != nil {
		return err
	}

	switch true {
	case i < 0xFD:
		*v = VarInt(i)
	case i == 0xFD:
		var j uint16
		if err := d.Decode(&j); err != nil {
			return err
		}
		*v = VarInt(j)
	case i == 0xFE:
		var j uint32
		if err := d.Decode(&j); err != nil {
			return err
		}
		*v = VarInt(j)
	case i == 0xFF:
		var j uint64
		if err := d.Decode(&j); err != nil {
			return err
		}
		*v = VarInt(j)
	}

	return nil
}

func (v VarInt) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	var i interface{}
	switch true {
	case v < 0xFD:
		i = uint8(v)
	case v <= 0xFFFF:
		buf.Write([]byte{0xFD})
		i = uint16(v)
	case v <= 0xFFFFFFFF:
		buf.Write([]byte{0xFE})
		i = uint32(v)
	default:
		buf.Write([]byte{0xFF})
		i = uint64(v)
	}
	b, err := binary.Marshal(i)
	if err != nil {
		return nil, err
	}

	if _, err = buf.Write(b); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type MsgHeaders struct {
	Count        VarInt
	BlockHeaders []BlockHeader
}

func (h *MsgHeaders) UnmarshalBinary(r io.Reader) error {
	d := binary.NewDecoder(r)
	if err := d.Decode(&h.Count); err != nil {
		return err
	}

	h.BlockHeaders = make([]BlockHeader, h.Count)
	for i := VarInt(0); i < h.Count; i++ {
		if i == VarInt(1840) {
			fmt.Println()
		}
		var bh BlockHeader
		if err := d.Decode(&bh); err != nil {
			return err
		}
		h.BlockHeaders[i] = bh
	}

	return nil
}

type BlockHeader struct {
	Version int32

	//The hash value of the previous block this particular block references
	PrevBlockHash [32]byte

	//The reference to a Merkle tree collection which is a hash of all transactions related to this block
	MerkleRoot [32]byte

	//A timestamp recording when this block was created (Will overflow in 2106[2])
	Timestamp uint32

	// The calculated difficulty target being used for this block
	Bits uint32

	// The nonce used to generate this block… to allow variations of the header and compute different hashes
	Nonce uint32

	// Number of transaction entries, this value is always 0
	TxnCount VarInt
}

// 1. Send getdata Message: Request full blocks based on the headers.

// 2. Receive Blocks: Process and store the blocks.

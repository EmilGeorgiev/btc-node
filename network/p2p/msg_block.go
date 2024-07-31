package p2p

import (
	"bytes"
	"crypto/sha256"
	"io"

	"github.com/EmilGeorgiev/btc-node/network/binary"
)

// MsgBlock represents a Bitcoin block, including its header and transactions.
type MsgBlock struct {
	BlockHeader
	Transactions []MsgTx
}

// MarshalBinary serializes the MsgBlock into a byte slice.
func (mb MsgBlock) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	b, err := binary.Marshal(mb.BlockHeader)
	if err != nil {
		return nil, err
	}

	if _, err = buf.Write(b); err != nil {
		return nil, err
	}

	for _, tx := range mb.Transactions {
		b, err = binary.Marshal(tx)
		if err != nil {
			return nil, err
		}
		if _, err = buf.Write(b); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary deserializes data from an io.Reader into the MsgBlock.
func (mb *MsgBlock) UnmarshalBinary(r io.Reader) error {
	d := binary.NewDecoder(r)
	if err := d.Decode(&mb.BlockHeader); err != nil {
		return err
	}

	mb.Transactions = make([]MsgTx, mb.BlockHeader.TxnCount)
	for i := VarInt(0); i < mb.BlockHeader.TxnCount; i++ {

		var tx MsgTx
		if err := d.Decode(&tx); err != nil {
			return err
		}
		mb.Transactions[i] = tx
	}

	return nil
}

// GetHash calculates and returns the double SHA-256 hash of the block header.
func (mb *MsgBlock) GetHash() [32]byte {
	b, _ := binary.Marshal(mb.BlockHeader)
	firstHash := sha256.Sum256(b[:80])
	return sha256.Sum256(firstHash[:])
}

// Reverse reverses a 32-byte array.
func Reverse(input [32]byte) []byte {
	l := len(input)
	reversed := make([]byte, l)
	for i, n := range input {
		j := l - i - 1
		reversed[j] = n
	}
	return reversed
}

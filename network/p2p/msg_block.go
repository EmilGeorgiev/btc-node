package p2p

import (
	"bytes"
	"crypto/sha256"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"io"
)

type MsgBlock struct {
	BlockHeader
	Transactions []MsgTx
}

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

func (mb *MsgBlock) GetHash() [32]byte {
	b, _ := binary.Marshal(mb.BlockHeader)
	firstHash := sha256.Sum256(b)
	return sha256.Sum256(firstHash[:])
}

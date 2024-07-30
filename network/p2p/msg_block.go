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

var count int

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

	//hash := mb.GetHash()
	//log.Printf("Block HASH-HASH-HASH-HASH that failed: %x\n The prev block hash is: %x\n", Reverse(hash[:]), Reverse(mb.PrevBlockHash[:]))

	//count++
	//if count == 20 {
	//	time.Sleep(10 * time.Second)
	//	panic("stop to see the blocks")
	//}
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
	firstHash := sha256.Sum256(b[:80])
	return sha256.Sum256(firstHash[:])
}

func Reverse(input [32]byte) []byte {
	l := len(input)
	reversed := make([]byte, l)
	for i, n := range input {
		j := l - i - 1
		reversed[j] = n
	}
	return reversed
}

//0000000068d61900e15a089c931e8365ae63f74fc1a6a246b6f1b3726fe28c0a
//000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f

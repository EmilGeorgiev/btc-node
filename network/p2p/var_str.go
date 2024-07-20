package p2p

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// VarStr ...
type VarStr struct {
	Length uint8
	String string
}

func (v *VarStr) UnmarshalBinary(r io.Reader) error {
	lengthBuf := make([]byte, 1)
	if _, err := r.Read(lengthBuf); err != nil {
		return fmt.Errorf("varStr.UnmarshalBinary: %+v", err)
	}
	v.Length = uint8(lengthBuf[0])

	stringBuf := make([]byte, v.Length)
	if _, err := r.Read(stringBuf); err != nil {
		return fmt.Errorf("varStr.UnmarshalBinary: %+v", err)
	}
	v.String = string(stringBuf)

	return nil
}

func NewVarStr(str string) VarStr {
	return VarStr{
		Length: uint8(len(str)), // TODO: implement var_int
		String: str,
	}
}

func checksum(data []byte) [checksumLength]byte {
	hash := sha256.Sum256(data)
	hash = sha256.Sum256(hash[:])
	var hashArr [checksumLength]byte
	copy(hashArr[:], hash[0:checksumLength])

	return hashArr
}

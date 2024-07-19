package p2p

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

// VarStr ...
type VarStr struct {
	Length uint8
	String string
}

func NewVarStr(str string) VarStr {
	return VarStr{
		Length: uint8(len(str)), // TODO: implement var_int
		String: str,
	}
}

// Serialize ...
func (v VarStr) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.LittleEndian, v.Length); err != nil {
		return nil, err
	}

	if _, err := buf.Write([]byte(v.String)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func checksum(data []byte) [checksumLength]byte {
	hash := sha256.Sum256(data)
	hash = sha256.Sum256(hash[:])
	var hashArr [checksumLength]byte
	copy(hashArr[:], hash[0:checksumLength])

	return hashArr
}

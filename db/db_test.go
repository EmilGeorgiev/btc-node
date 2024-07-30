package db

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

func TestBlockRepo_Savek(t *testing.T) {
	dbPath := "/tmp/my.db"
	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewBlockRepo(db.DB)
	require.NoError(t, err)

	block, err := repo.GetLast()
	fmt.Println(err)
	fmt.Printf("prev block hah: %x\n", block.PrevBlockHash)
	hash := block.GetHash()
	fmt.Printf("block hash: %x\n", p2p.Reverse(hash[:]))
}

// Test Functions
func TestBlockRepo_SaveAndGetBlock(t *testing.T) {
	dbPath := t.TempDir() + "/blocks.db"
	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewBlockRepo(db.DB)
	require.NoError(t, err)

	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	block := newMsgBlock(blockHash)

	err = repo.Save(block)
	require.NoError(t, err)

	hash := block.GetHash()
	actual, err := repo.Get(hash)
	require.NoError(t, err)

	require.Equal(t, block, actual)
}

func TestBlockRepo_GetLastBlockWhenSaveOneBlock(t *testing.T) {
	dbPath := t.TempDir() + "/blocks.db"
	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewBlockRepo(db.DB)
	require.NoError(t, err)

	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	block := newMsgBlock(blockHash)

	err = repo.Save(block)
	require.NoError(t, err)

	actual, err := repo.GetLast()
	require.NoError(t, err)
	require.Equal(t, block, actual)
}

func TestBlockRepo_GetLastBlockWhenSaveMultipleBlocks(t *testing.T) {
	dbPath := t.TempDir() + "/blocks.db"
	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewBlockRepo(db.DB)
	require.NoError(t, err)

	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	block := newMsgBlock(blockHash)
	block2 := newMsgBlock(block.GetHash())
	block3 := newMsgBlock(block2.GetHash())
	block4 := newMsgBlock(block3.GetHash())

	err = repo.Save(block)
	require.NoError(t, err)
	err = repo.Save(block2)
	require.NoError(t, err)
	err = repo.Save(block3)
	require.NoError(t, err)
	err = repo.Save(block4)
	require.NoError(t, err)

	actual, err := repo.GetLast()
	require.NoError(t, err)
	require.Equal(t, block4, actual)
}

func newMsgBlock(prevBlockHash [32]byte) p2p.MsgBlock {

	return p2p.MsgBlock{
		BlockHeader: p2p.BlockHeader{
			Version:       1,
			PrevBlockHash: prevBlockHash,
			MerkleRoot:    [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F, 0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A},
			Timestamp:     1721836804,
			Bits:          1721836804,
			Nonce:         1721836804,
			TxnCount:      p2p.VarInt(1),
		},
	}
}

func reverseBytes(data []byte) []byte {
	length := len(data)
	reversed := make([]byte, length)
	for i := 0; i < length; i++ {
		reversed[i] = data[length-1-i]
	}
	return reversed
}

func TestBlockRepo_GetLaock(t *testing.T) {
	dbPath := "/tmp/my.db"
	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewBlockRepo(db.DB)
	require.NoError(t, err)

	actual, err := repo.GetLast()
	require.NoError(t, err)

	h := actual.GetHash()
	fmt.Printf("1Last block in DB is %x\n", p2p.Reverse(h[:]))

	block, err := repo.Get(actual.PrevBlockHash)
	require.NoError(t, err)
	h = block.GetHash()
	fmt.Printf("2 Previous block is: %x\n", p2p.Reverse(h[:]))

	block, err = repo.Get(block.PrevBlockHash)
	require.NoError(t, err)
	h = block.GetHash()
	fmt.Printf("3 Previous block is: %x\n", p2p.Reverse(h[:]))

	block, err = repo.Get(block.PrevBlockHash)
	require.NoError(t, err)
	h = block.GetHash()
	fmt.Printf("4 Previous block is: %x\n", p2p.Reverse(h[:]))
}

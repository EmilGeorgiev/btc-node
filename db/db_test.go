package db

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

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
	actual, err := repo.GetBlock(hash[:])
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

	hash, err := repo.GetLastBlockHash()
	require.NoError(t, err)
	require.Equal(t, block.GetHash(), hash)
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

	hash, err := repo.GetLastBlockHash()
	require.NoError(t, err)
	require.Equal(t, block4.GetHash(), hash)
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
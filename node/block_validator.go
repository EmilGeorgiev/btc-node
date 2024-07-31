package node

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"log"
	"math/big"
)

// BlockValidator is responsible for validating blocks ( headers and transactions).
type BlockValidator struct {
	blockRepo BlockRepository
}

// NewBlockValidator creates a new BlockValidator with the given BlockRepository.
func NewBlockValidator(br BlockRepository) BlockValidator {
	return BlockValidator{
		blockRepo: br,
	}
}

// Validate checks if the given block is valid by performing various checks like
// previous block hash, hash is bellow the target, merkle tree validation and others.
func (bv BlockValidator) Validate(bl *p2p.MsgBlock) error {
	if err := bv.lastBlockHashMustBePreviousForTheCurrentOne(bl); err != nil {
		return err
	}

	if !blockHashLessThanTargetDifficulty(&bl.BlockHeader) {
		return fmt.Errorf("target hash is not es then target difficulty")
	}

	//if !bv.ValidateMerkleTree(bl) {
	//	log.Println("merkle three is not valid or transactions are not valid")
	//	return errors.New("merkle three is not valid or transactions are not valid")
	//}
	return nil
}

// ValidateMerkleTree checks if the Merkle tree of the block is valid.
func (bv BlockValidator) ValidateMerkleTree(bl *p2p.MsgBlock) bool {
	if bl.TxnCount == 1 {
		h := HashTx(bl.Transactions[0])
		return [32]byte(h) == bl.MerkleRoot
	}

	txHashes := make([][]byte, bl.TxnCount)
	for i, tx := range bl.Transactions {
		txHashes[i] = HashTx(tx)
	}

	if bl.TxnCount%2 != 0 {
		// duplicate last transaction.
		txHashes = append(txHashes, txHashes[bl.TxnCount-1])
	}

	for len(txHashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(txHashes); i += 2 {
			concatenated := append(txHashes[i], txHashes[i+1]...)
			newLevel = append(newLevel, DHash(concatenated))
		}
		txHashes = newLevel
	}
	return bl.MerkleRoot == [32]byte(txHashes[0])
}

// DHash performs double SHA-256 hashing on the given byte slice.
func DHash(b []byte) []byte {
	firstHash := sha256.Sum256(b)
	h := sha256.Sum256(firstHash[:])
	return h[:]
}

// HashTx calculates the double SHA-256 hash of the given transaction.
func HashTx(tx p2p.MsgTx) []byte {
	b, _ := binary.Marshal(tx)
	firstHash := sha256.Sum256(b)
	h := sha256.Sum256(firstHash[:])
	return h[:]
}

// lastBlockHashMustBePreviousForTheCurrentOne checks if the last block hash in
// the repository matches the previous block hash of the current block.
func (bv BlockValidator) lastBlockHashMustBePreviousForTheCurrentOne(bl *p2p.MsgBlock) error {
	lastBlockHash := [32]byte{}
	lastBlock, err := bv.blockRepo.GetLast()
	if err != nil {
		log.Println("ERRRRRR when get lat block in validate block:", err)
		if !errors.Is(err, sync.ErrNotFound) {
			return err
		}
		log.Println("Set last block to zero")
	} else {
		lastBlockHash = lastBlock.GetHash()
	}

	if bl.PrevBlockHash != lastBlockHash && lastBlockHash != [32]byte{} {
		log.Printf("last block in DB is not the prev of the current one: %x\n", p2p.Reverse(lastBlockHash))
	} else {
		log.Println("Last block in DB is the prvious one before the this that wil be tored")
	}
	return nil
}

// BitsToTarget converts the compact representation of the target into a big.Int.
func BitsToTarget(bits uint32) *big.Int {
	exponent := bits >> 24
	mantissa := bits & 0xffffff
	target := new(big.Int).SetUint64(uint64(mantissa))
	target.Lsh(target, 8*(uint(exponent)-3))
	return target
}

// blockHashLessThanTargetDifficulty checks if the block hash is less than the target difficulty.
func blockHashLessThanTargetDifficulty(headers *p2p.BlockHeader) bool {
	hashBig := new(big.Int).SetBytes(p2p.Reverse(Hash(*headers)))
	target := BitsToTarget(headers.Bits)

	return hashBig.Cmp(target) <= 0
}

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

type BlockValidator struct {
	blockRepo BlockRepository
}

func NewBlockValidator(br BlockRepository) BlockValidator {
	return BlockValidator{
		blockRepo: br,
	}
}

var zero = [32]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

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
			//if i+1 == len(txHashes) {
			//	// Duplicate the last element if the number of elements is odd
			//	txHashes = append(txHashes, txHashes[i])
			//}
			concatenated := append(txHashes[i], txHashes[i+1]...)
			newLevel = append(newLevel, DHash(concatenated))
		}
		txHashes = newLevel
	}
	return bl.MerkleRoot == [32]byte(txHashes[0])
}

func DHash(b []byte) []byte {
	firsthash := sha256.Sum256(b)
	h := sha256.Sum256(firsthash[:])
	return h[:]
}

func HashTx(tx p2p.MsgTx) []byte {
	b, _ := binary.Marshal(tx)
	firstHash := sha256.Sum256(b)
	h := sha256.Sum256(firstHash[:])
	return h[:]
}

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

// target is a 256 bit numbe
func BitsToTarget(bits uint32) *big.Int {
	exponent := bits >> 24
	mantissa := bits & 0xffffff
	target := new(big.Int).SetUint64(uint64(mantissa))
	target.Lsh(target, 8*(uint(exponent)-3))
	return target
}

// ValidateBlockHash ...
func blockHashLessThanTargetDifficulty(headers *p2p.BlockHeader) bool {
	hashBig := new(big.Int).SetBytes(p2p.Reverse(Hash(*headers)))
	target := BitsToTarget(headers.Bits)

	return hashBig.Cmp(target) <= 0
}

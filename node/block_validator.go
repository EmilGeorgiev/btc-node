package node

import (
	"errors"
	"fmt"
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
	return nil
}

func (bv BlockValidator) lastBlockHashMustBePreviousForTheCurrentOne(bl *p2p.MsgBlock) error {
	var lastBlockHash [32]byte
	lastBlock, err := bv.blockRepo.GetLast()
	if err != nil {
		if !errors.Is(err, sync.ErrNotFound) {
			return err
		}
		lastBlockHash = zero
	} else {
		lastBlockHash = lastBlock.GetHash()
	}

	if bl.PrevBlockHash != lastBlockHash {
		log.Printf("last block in DB is not the prev of the current one: %x\n", p2p.Reverse(lastBlockHash[:]))
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
	hash := Hash(*headers)
	hashBig := new(big.Int).SetBytes(hash[:])
	target := BitsToTarget(headers.Bits)

	return hashBig.Cmp(target) <= 0
}

package db

import (
	"encoding/json"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	bolt "go.etcd.io/bbolt"
	"log"
)

var (
	blockBucket      = []byte("BlockBucket")
	lastBlockBucket  = []byte("LastBlockBucket")
	prevToNextBucket = []byte("PrevToNextBucket")
	lastBlockKey     = []byte("LastBlockKey")
)

type BlocksRepo struct {
	db *bolt.DB
}

func NewBlockRepo(db *bolt.DB) (*BlocksRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(blockBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(lastBlockBucket); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(prevToNextBucket); err != nil {
			return err
		}
		return nil
	})

	return &BlocksRepo{db}, err
}

func (db *BlocksRepo) Save(block p2p.MsgBlock) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		blockBkt := tx.Bucket(blockBucket)
		hash := block.GetHash()
		data, _ := json.Marshal(block)
		err := blockBkt.Put(hash[:], data)
		if err != nil {
			return err
		}

		prevToNext := tx.Bucket(prevToNextBucket)
		err = prevToNext.Put(block.PrevBlockHash[:], hash[:])
		if err != nil {
			return err
		}

		lastBlock := tx.Bucket(lastBlockBucket)
		return lastBlock.Put(lastBlockKey, hash[:])
	})
}

func (db *BlocksRepo) GetLast() (p2p.MsgBlock, error) {
	var block p2p.MsgBlock
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(lastBlockBucket)
		lastBlockHash := b.Get(lastBlockKey)

		log.Printf("check whether the block hash is realy thelast one %x\n", p2p.Reverse(lastBlockHash[:]))
		prevToNext := tx.Bucket(prevToNextBucket)
		for {
			nextBlockHash := prevToNext.Get(lastBlockHash)
			if len(nextBlockHash) != 0 {
				lastBlockHash = nextBlockHash
				continue
			}
			break
		}
		log.Printf("after the check, the real last block is %x\n", p2p.Reverse(lastBlockHash[:]))

		bl := tx.Bucket(blockBucket)
		data := bl.Get(lastBlockHash)
		if b == nil || len(data) == 0 {
			return sync.ErrNotFound
		}

		return json.Unmarshal(data, &block)
	})
	return block, err
}

func (db *BlocksRepo) Get(hash [32]byte) (p2p.MsgBlock, error) {
	var block p2p.MsgBlock
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blockBucket)
		data := b.Get(hash[:])
		if data == nil {
			return fmt.Errorf("block not found")
		}
		return json.Unmarshal(data, &block)
	})
	return block, err
}

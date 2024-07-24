package db

import (
	"context"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	bolt "go.etcd.io/bbolt"
	"os"
)

type DbImpl struct {
	db *bolt.DB
}

type Config struct {
	Path string
	Mode os.FileMode
}

func NewDB() (DbImpl, error) {
	db, err := bolt.Open("blocks.db", 0600, nil)
	if err != nil {
		return DbImpl{}, err
	}

	return DbImpl{db}, err
}

func (r DbImpl) Store(ctx context.Context, block p2p.MsgBlock) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("blocks"))
		if err != nil {
			return err
		}

		blockHash := block.GetHash()
		blockBytes, err := block.MarshalBinary()
		return bucket.Put(blockHash[:], blockBytes)
	})
}

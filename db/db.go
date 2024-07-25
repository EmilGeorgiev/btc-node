package db

import bolt "go.etcd.io/bbolt"

type BoltDB struct {
	*bolt.DB
}

func NewBoltDB(dbPath string) (BoltDB, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return BoltDB{}, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists(blockBucket); err != nil {
			return err
		}

		if _, err = tx.CreateBucketIfNotExists(lastBlockBucket); err != nil {
			return err
		}

		if _, err = tx.CreateBucketIfNotExists(prevToNextBucket); err != nil {
			return err
		}
		return nil
	})

	return BoltDB{db}, err
}

func (db BoltDB) Close() {
	db.DB.Close()
}

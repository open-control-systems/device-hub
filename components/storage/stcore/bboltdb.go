package stcore

import (
	"github.com/open-control-systems/device-hub/components/status"
	"go.etcd.io/bbolt"
)

// NewBboltDB initialization.
//
// Parameters:
//   - dbPath - database file path, if it doesn't exist then it will be created automatically.
//
// References:
//   - https://github.com/etcd-io/bbolt
func NewBboltDB(dbPath string, opts *bbolt.Options) (*bbolt.DB, error) {
	db, err := bbolt.Open(dbPath, 0600, opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// BboltDBBucket is a wrapper over the bbolt database to operate on a single bucket.
type BboltDBBucket struct {
	db     *bbolt.DB
	bucket string
}

// NewBboltDBBucket initialization.
//
// Parameters:
//   - db - bbolt database instance.
//   - bucket - bbolt database bucket.
func NewBboltDBBucket(db *bbolt.DB, bucket string) *BboltDBBucket {
	return &BboltDBBucket{
		db:     db,
		bucket: bucket,
	}
}

// Read reads data from the database bucket.
func (b *BboltDBBucket) Read(key string) ([]byte, error) {
	var ret []byte

	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return status.StatusNoData
		}

		ret = bucket.Get([]byte(key))
		if ret == nil {
			return status.StatusNoData
		}

		return nil
	})
	if err != nil {
		return []byte{}, err
	}

	return ret, nil
}

// Write writes data to the database bucket.
func (b *BboltDBBucket) Write(key string, value []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.bucket))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), value)
	})
}

// Remove removes data from the database bucket.
func (b *BboltDBBucket) Remove(key string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(key))
	})
}

// ForEach iterates over all key-value pairs in the database bucket.
func (b *BboltDBBucket) ForEach(fn func(string, []byte) error) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			return fn(string(k), v)
		})
	})
}

// Close is non-operational.
func (*BboltDBBucket) Close() error {
	return nil
}

package stcore

import (
	"fmt"

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

// Read reads a blob of data from bbolt database.
func (b *BboltDBBucket) Read(key string) (Blob, error) {
	blob := Blob{}

	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return status.StatusNoData
		}

		data := bucket.Get([]byte(key))
		if data == nil {
			return status.StatusNoData
		}

		blob.Data = data

		return nil
	})
	if err != nil {
		return Blob{}, err
	}

	return blob, nil
}

// Write write a blob to the database bucket.
func (b *BboltDBBucket) Write(key string, blob Blob) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.bucket))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), blob.Data)
	})
}

// Remove removes a blob from the database bucket.
func (b *BboltDBBucket) Remove(key string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(key))
	})
}

// ForEach iterates over all blobs in the database bucket.
func (b *BboltDBBucket) ForEach(fn func(key string, b Blob) error) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return fmt.Errorf("bucket=%s not found", b.bucket)
		}

		return bucket.ForEach(func(k, v []byte) error {
			return fn(string(k), Blob{Data: v})
		})
	})
}

// Close is non-operational.
func (*BboltDBBucket) Close() error {
	return nil
}

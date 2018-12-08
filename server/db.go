package server

import (
	"os"
	"path/filepath"
	"time"

	bolt "github.com/coreos/bbolt"
)

type localDB struct {
	path string
	db   *bolt.DB
}

type KeyValue struct {
	Key   string
	Value []byte
}

func NewLocalDB(path string) (*localDB, error) {
	base := filepath.Dir(path)
	if err := os.MkdirAll(base, 0700); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second * 1})
	if err != nil {
		return nil, err
	}

	return &localDB{
		path: path,
		db:   db,
	}, nil
}

func (l *localDB) CreateBucket(bucket string) error {
	return l.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
			return err
		}
		return nil
	})
}

func (l *localDB) Set(bucket, key string, value []byte) error {
	return l.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		if err := b.Put([]byte(key), value); err != nil {
			return err
		}
		return nil
	})
}

func (l *localDB) Get(bucket, key string) (*KeyValue, error) {
	var data *KeyValue
	if err := l.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		data = &KeyValue{
			Key:   key,
			Value: v,
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return data, nil
}

func (l *localDB) GetAll(bucket string) ([]*KeyValue, error) {
	var data []*KeyValue
	if err := l.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			data = append(data, &KeyValue{
				Key:   string(k),
				Value: v,
			})
		}

		return nil

	}); err != nil {
		return nil, err
	}

	return data, nil
}

package datastore

import (
	"encoding/json"
	"time"

	"github.com/containerd/containerd/errdefs"
	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	pruneTimeout = time.Second * 90
)

func (s *service) prune() error {
	s.lock.Lock()
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.dsTombstoneBucketName))
		if b == nil {
			return errdefs.ToGRPC(errors.Wrapf(errdefs.ErrNotFound, "bucket %s", s.dsTombstoneBucketName))
		}
		if err := b.ForEach(func(k, v []byte) error {
			var t *tombstone
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			if time.Now().After(t.Timestamp.Add(pruneTimeout)) {
				logrus.Debugf("prune: removing expired tombstone record %s", k)
				if err := pruneRemove(b, k); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	s.lock.Unlock()

	return nil
}

func pruneRemove(b *bolt.Bucket, key []byte) error {
	return b.Delete(key)
}

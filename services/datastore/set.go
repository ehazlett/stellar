package datastore

import (
	"context"

	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Set(ctx context.Context, req *api.SetRequest) (*ptypes.Empty, error) {
	var err error
	s.lock.Lock()
	err = s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(req.Bucket))
		if err != nil {
			return err
		}
		if err := b.Put([]byte(req.Key), req.Value); err != nil {
			return err
		}
		return nil
	})
	// check for any tombstone records and remove if exists as this is an updated record
	err = s.db.Update(func(tx *bolt.Tx) error {
		tb := tx.Bucket([]byte(s.dsTombstoneBucketName))
		if tb == nil {
			// no tombstone bucket; ignore
			return nil
		}
		tsKey := s.tombstoneKeyName(req.Bucket, req.Key)
		if v := tb.Get([]byte(tsKey)); v != nil {
			// ignore delete errors
			if err := tb.Delete([]byte(tsKey)); err != nil {
				return err
			}
			logrus.WithField("key", tsKey).Debug("removed tombstone record as key was updated")
		}
		return nil
	})
	s.lock.Unlock()

	logrus.WithFields(logrus.Fields{
		"bucket": req.Bucket,
		"key":    req.Key,
		"sync":   req.Sync,
	}).Debug("updated datastore")

	if req.Sync {
		if err := s.replicateToPeers(ctx); err != nil {
			return empty, err
		}
	}

	return empty, err
}

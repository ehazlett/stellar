package datastore

import (
	"context"

	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Delete(ctx context.Context, req *api.DeleteRequest) (*ptypes.Empty, error) {
	s.lock.Lock()
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(req.Bucket))
		if b == nil {
			return ErrBucketDoesNotExist
		}
		if err := b.Delete([]byte(req.Key)); err != nil {
			return err
		}
		return nil
	})
	s.lock.Unlock()

	logrus.WithFields(logrus.Fields{
		"bucket": req.Bucket,
		"key":    req.Key,
	}).Debug("removed key from datastore")

	if err == nil && req.Sync {
		// sync to peers
		if err := s.syncWithPeers(ctx, &syncOp{
			action: syncDelete,
			bucket: req.Bucket,
			key:    req.Key,
		}); err != nil {
			return &ptypes.Empty{}, err
		}
	}

	return &ptypes.Empty{}, err
}

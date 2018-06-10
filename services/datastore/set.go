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

	s.lock.Unlock()

	logrus.WithFields(logrus.Fields{
		"bucket": req.Bucket,
		"key":    req.Key,
	}).Debug("updated datastore")

	if err == nil && req.Sync {
		// sync to peers
		if err := s.syncWithPeers(ctx, &syncOp{
			action: syncSet,
			bucket: req.Bucket,
			key:    req.Key,
			value:  req.Value,
		}); err != nil {
			return &ptypes.Empty{}, err
		}
	}

	return &ptypes.Empty{}, err
}

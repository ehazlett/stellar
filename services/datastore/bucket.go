package datastore

import (
	"context"

	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) CreateBucket(ctx context.Context, req *api.CreateBucketRequest) (*ptypes.Empty, error) {
	var err error
	s.lock.Lock()
	err = s.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(req.Bucket)); err != nil {
			return err
		}
		return nil
	})
	s.lock.Unlock()

	return empty, err
}

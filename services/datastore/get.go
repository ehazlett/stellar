package datastore

import (
	"context"

	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
)

func (s *service) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	var val []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(req.Bucket))
		if b == nil {
			return ErrBucketDoesNotExist
		}
		val = b.Get([]byte(req.Key))
		return nil
	})
	return &api.GetResponse{
		Bucket: req.Bucket,
		Value:  val,
	}, err
}

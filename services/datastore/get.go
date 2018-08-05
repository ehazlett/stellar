package datastore

import (
	"context"

	"github.com/containerd/containerd/errdefs"
	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/pkg/errors"
)

func (s *service) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	var val []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(req.Bucket))
		if b == nil {
			return errdefs.ToGRPC(errors.Wrapf(errdefs.ErrNotFound, "bucket %s", req.Bucket))
		}
		val = b.Get([]byte(req.Key))
		return nil
	})
	return &api.GetResponse{
		Bucket: req.Bucket,
		Data: &api.KeyValue{
			Key:   req.Key,
			Value: val,
		},
	}, err
}

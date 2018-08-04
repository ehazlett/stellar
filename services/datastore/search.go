package datastore

import (
	"bytes"
	"context"

	"github.com/containerd/containerd/errdefs"
	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/pkg/errors"
)

func (s *service) Search(ctx context.Context, req *api.SearchRequest) (*api.SearchResponse, error) {
	var data []*api.KeyValue

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(req.Bucket))
		if b == nil {
			return errdefs.ToGRPC(errors.Wrapf(errdefs.ErrNotFound, "bucket %s", req.Bucket))
		}

		c := b.Cursor()

		prefix := []byte(req.Prefix)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			data = append(data, &api.KeyValue{
				Key:   string(k),
				Value: v,
			})
		}
		return nil
	})

	return &api.SearchResponse{
		Bucket: req.Bucket,
		Data:   data,
	}, err
}

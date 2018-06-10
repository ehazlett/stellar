package datastore

import (
	"bytes"
	"context"

	bolt "github.com/coreos/bbolt"
	api "github.com/ehazlett/stellar/api/services/datastore/v1"
)

func (s *service) Backup(ctx context.Context, _ *api.BackupRequest) (*api.BackupResponse, error) {
	buf := bytes.NewBuffer(nil)
	err := s.db.View(func(tx *bolt.Tx) error {
		if _, err := tx.WriteTo(buf); err != nil {
			return err
		}

		return nil
	})

	return &api.BackupResponse{
		Data: buf.Bytes(),
	}, err
}

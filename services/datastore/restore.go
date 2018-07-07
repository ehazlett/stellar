package datastore

import (
	"context"
	"os"

	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Restore(ctx context.Context, req *api.RestoreRequest) (*ptypes.Empty, error) {
	s.lock.Lock()
	s.db.Close()
	defer s.lock.Unlock()

	if err := os.Remove(dbFilename); err != nil {
		return empty, err
	}

	f, err := os.OpenFile(dbFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return empty, err
	}
	f.Write(req.Data)
	f.Close()

	db, err := s.openDB()
	if err != nil {
		return empty, err
	}

	s.db = db

	return empty, err
}

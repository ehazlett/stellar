package datastore

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Shutdown(ctx context.Context, req *api.ShutdownRequest) (*ptypes.Empty, error) {
	logrus.Debug("performing shutdown sync with peers")
	if err := s.replicateToPeers(ctx); err != nil {
		return empty, err
	}
	return empty, nil
}

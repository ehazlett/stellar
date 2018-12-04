package server

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	api "github.com/stellarproject/radiant/api/v1"
)

func (s *Server) AddServer(ctx context.Context, req *api.AddServerRequest) (*ptypes.Empty, error) {
	if err := s.datastore.Add(req.Server.Host, req.Server); err != nil {
		return empty, err
	}
	return empty, nil
}

package server

import (
	"context"

	api "github.com/ehazlett/blackbird/api/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *Server) AddServer(ctx context.Context, req *api.AddServerRequest) (*ptypes.Empty, error) {
	if err := s.datastore.Add(req.Server.Host, req.Server); err != nil {
		return empty, err
	}
	return empty, nil
}

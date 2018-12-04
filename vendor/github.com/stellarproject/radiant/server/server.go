package server

import (
	"context"

	api "github.com/stellarproject/radiant/api/v1"
)

func (s *Server) Servers(ctx context.Context, req *api.ServersRequest) (*api.ServersResponse, error) {
	servers, err := s.datastore.Servers()
	if err != nil {
		return nil, err
	}
	return &api.ServersResponse{
		Servers: servers,
	}, nil
}

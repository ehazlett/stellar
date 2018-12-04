package server

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	api "github.com/stellarproject/radiant/api/v1"
)

func (s *Server) RemoveServer(ctx context.Context, req *api.RemoveServerRequest) (*ptypes.Empty, error) {
	if err := s.datastore.Remove(req.Host); err != nil {
		return empty, err
	}
	return empty, nil
}

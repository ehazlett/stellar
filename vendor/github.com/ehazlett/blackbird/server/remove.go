package server

import (
	"context"

	api "github.com/ehazlett/blackbird/api/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *Server) RemoveServer(ctx context.Context, req *api.RemoveServerRequest) (*ptypes.Empty, error) {
	if err := s.datastore.Remove(req.Host); err != nil {
		return empty, err
	}
	return empty, nil
}

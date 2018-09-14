package server

import (
	"context"

	api "github.com/ehazlett/blackbird/api/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *Server) Reload(ctx context.Context, req *api.ReloadRequest) (*ptypes.Empty, error) {
	caddyfile, err := s.getCaddyConfig()
	if err != nil {
		return empty, err
	}

	s.instance, err = s.instance.Restart(caddyfile)
	if err != nil {
		logrus.Error(string(caddyfile.Body()))
		return empty, err
	}

	logrus.Info("proxy reloaded")
	return empty, nil
}

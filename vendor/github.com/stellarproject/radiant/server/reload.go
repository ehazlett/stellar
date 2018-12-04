package server

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	api "github.com/stellarproject/radiant/api/v1"
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

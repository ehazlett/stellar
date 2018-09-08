package proxy

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Reload(ctx context.Context, req *api.ReloadRequest) (*ptypes.Empty, error) {
	return empty, s.reload()
}

func (s *service) reload() error {
	caddyfile, err := s.getCaddyConfig()
	if err != nil {
		return err
	}

	id, err := generateID(caddyfile.Body())
	if err != nil {
		return err
	}

	if id == s.configID {
		return nil
	}

	s.configID = id
	logrus.Debugf("proxy: reloading with config id=%s", s.configID)

	s.instance, err = s.instance.Restart(caddyfile)
	return err
}

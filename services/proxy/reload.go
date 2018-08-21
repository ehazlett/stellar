package proxy

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Reload(ctx context.Context, req *api.ReloadRequest) (*ptypes.Empty, error) {
	if err := s.reload(); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *service) reload() error {
	c, err := s.client()
	if err != nil {
		return err
	}
	defer c.Close()

	apps, err := c.Application().List()
	if err != nil {
		return err
	}

	if err := s.proxy.Reload(apps); err != nil {
		return nil
	}

	return nil

}

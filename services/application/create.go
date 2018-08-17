package application

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Create(ctx context.Context, req *api.CreateRequest) (*ptypes.Empty, error) {
	logrus.Debugf("creating application %s", req.Name)
	c, err := s.client()
	if err != nil {
		return empty, err
	}
	defer c.Close()

	for _, service := range req.Services {
		// TODO: make cluster aware
		if err := c.Node().CreateContainer(req.Name, service); err != nil {
			return empty, err
		}
	}
	return empty, nil
}

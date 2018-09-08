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

	nodes, err := c.Cluster().Nodes()
	if err != nil {
		return empty, err
	}

	nodeIdx := 0
	for _, service := range req.Services {
		// get random peer for deploy
		node := nodes[nodeIdx]
		nc, err := s.nodeClient(node.Name)
		if err != nil {
			return empty, err
		}

		if err := nc.Node().CreateContainer(req.Name, service); err != nil {
			return empty, err
		}

		// update proxy
		if err := nc.Proxy().Reload(); err != nil {
			return empty, err
		}

		nc.Close()

		// update peer index for next deploy
		nodeIdx++
		if nodeIdx >= len(nodes) {
			nodeIdx = 0
		}
	}
	return empty, nil
}

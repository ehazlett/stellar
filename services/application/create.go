package application

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Create(ctx context.Context, req *api.CreateRequest) (*ptypes.Empty, error) {
	appName := getAppName(req.Name)
	logrus.Debugf("creating application %s", appName)
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return empty, err
	}
	defer c.Close()

	nodes, err := c.Cluster().Nodes()
	if err != nil {
		return empty, err
	}

	// only create missing services
	containers, err := s.getApplicationContainers(appName)
	if err != nil {
		return empty, err
	}

	ids := map[string]struct{}{}

	for _, cnt := range containers {
		ids[cnt.Container.ID] = struct{}{}
	}

	services := []*nodeapi.Service{}
	for _, service := range req.Services {
		if _, ok := ids[req.Name+"."+service.Name]; ok {
			continue
		}

		services = append(services, service)
	}

	nodeIdx := 0
	for _, service := range services {
		// get random peer for deploy
		node := nodes[nodeIdx]
		nc, err := s.client(node.Address)
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

package application

import (
	"context"
	"fmt"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
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

	services := []*runtimeapi.Service{}
	for i, service := range req.Services {
		id := fmt.Sprintf("%s.%s.%d", req.Name, service.Name, i)
		if _, ok := ids[id]; ok {
			continue
		}

		services = append(services, service)
	}

	for _, service := range services {
		// get list of target nodes for the service
		scheduledNodes, err := c.Scheduler().Schedule(service, nodes)
		if err != nil {
			return empty, err
		}
		logrus.WithFields(logrus.Fields{
			"service": service.Name,
			"nodes":   scheduledNodes,
		}).Debug("scheduled nodes for service")
		for i, node := range scheduledNodes {
			nc, err := s.client(node.Address)
			if err != nil {
				return empty, err
			}

			// inject replica id into service name
			id := fmt.Sprintf("%s.%d", service.Name, i)

			if err := nc.Node().CreateContainer(req.Name, service, id); err != nil {
				return empty, err
			}

			// update proxy
			if err := nc.Proxy().Reload(); err != nil {
				return empty, err
			}

			nc.Close()
		}
	}

	if err := s.publish(&UpdateEvent{
		Application: req.Name,
		Action:      "create",
	}); err != nil {
		return empty, err
	}
	return empty, nil
}

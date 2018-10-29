package application

import (
	"context"
	"errors"
	"strings"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrApplicationNotFound = errors.New("application not found")
)

func (s *service) Delete(ctx context.Context, req *api.DeleteRequest) (*ptypes.Empty, error) {
	c, err := s.client(s.agent.Self().Address)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	appName := getAppName(req.Name)
	containers, err := s.getApplicationContainers(appName)
	if err != nil {
		return empty, err
	}

	if len(containers) == 0 {
		return empty, status.Errorf(codes.NotFound, "application %s not found", appName)
	}

	for _, cc := range containers {
		id := cc.Container.ID
		if !strings.HasPrefix(id, req.Name) {
			continue
		}
		logrus.Debugf("app delete: deleting container %s", id)
		nc, err := s.client(cc.Node.Address)
		if err != nil {
			logrus.Warnf("delete: error getting client for node %s: %s", cc.Node.ID, err)
			continue
		}

		if err := nc.Node().DeleteContainer(cc.Container.ID); err != nil {
			logrus.Warnf("delete: error deleting service on node %s: %s", cc.Node.ID, err)
			continue
		}

		name := id + ".stellar"
		if err := c.Nameserver().Delete("A", name); err != nil {
			return nil, err
		}

		nc.Close()
	}

	if err := s.publish(&UpdateEvent{
		Application: req.Name,
		Action:      "delete",
	}); err != nil {
		return empty, err
	}

	return empty, nil
}

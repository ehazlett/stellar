package application

import (
	"context"
	"strings"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) Restart(ctx context.Context, req *api.RestartRequest) (*ptypes.Empty, error) {
	appName := getAppName(req.Name)
	containers, err := s.getApplicationContainers(appName)
	if err != nil {
		return empty, err
	}

	if len(containers) == 0 {
		return empty, status.Errorf(codes.NotFound, "application %s not found", appName)
	}

	for _, cc := range containers {
		if !strings.HasPrefix(cc.Container.ID, req.Name) {
			continue
		}
		logrus.Debugf("restarting container %s on node %s", cc.Container.ID, cc.Node.Name)
		nc, err := s.nodeClient(cc.Node.Name)
		if err != nil {
			logrus.Warnf("delete: error getting client for node %s: %s", cc.Node.Name, err)
			continue
		}

		if err := nc.Node().RestartContainer(cc.Container.ID); err != nil {
			logrus.Warnf("restart: error restarting service on node %s: %s", cc.Node.Name, err)
			continue
		}

		nc.Close()
	}

	return empty, err
}

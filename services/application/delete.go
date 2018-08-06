package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/ehazlett/stellar"
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
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	containers, err := c.Cluster().Containers(fmt.Sprintf("labels.\"%s\"==\"%s\"", stellar.StellarApplicationLabel, req.Name))
	if err != nil {
		return empty, err
	}

	if len(containers) == 0 {
		return empty, status.Errorf(codes.NotFound, "application %s not found", req.Name)
	}

	for _, cc := range containers {
		id := cc.Container.ID
		logrus.Debugf("app delete: deleting container %s", id)
		nc, err := s.nodeClient(cc.Node.Name)
		if err != nil {
			logrus.Warnf("delete: error getting client for node %s: %s", cc.Node.Name, err)
			continue
		}

		if err := nc.Node().DeleteContainer(cc.Container.ID); err != nil {
			logrus.Warnf("delete: error deleting service on node %s: %s", cc.Node.Name, err)
			continue
		}

		name := id + ".stellar"
		if err := c.Nameserver().Delete("A", name); err != nil {
			return nil, err
		}

		nc.Close()
	}

	return empty, nil
}

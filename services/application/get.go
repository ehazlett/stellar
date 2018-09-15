package application

import (
	"context"
	"fmt"

	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *service) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	app, err := s.getApp(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &api.GetResponse{
		Application: app,
	}, nil
}

func (s *service) getApp(ctx context.Context, name string) (*api.App, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	containers, err := c.Cluster().Containers(fmt.Sprintf("labels.\"%s\"==\"%s\"", stellar.StellarApplicationLabel, name))
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, status.Errorf(codes.NotFound, "application %s not found", name)
	}

	app := &api.App{
		Name: name,
	}
	for _, container := range containers {
		svc, err := s.containerToService(ctx, container)
		if err != nil {
			logrus.Error(err)
			continue
		}
		app.Services = append(app.Services, svc)
	}

	return app, nil
}

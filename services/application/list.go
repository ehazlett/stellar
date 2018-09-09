package application

import (
	"context"
	"fmt"
	"sort"

	"github.com/ehazlett/stellar"
	api "github.com/ehazlett/stellar/api/services/application/v1"
	"github.com/sirupsen/logrus"
)

type AppSorter []*api.App

func (a AppSorter) Len() int           { return len(a) }
func (a AppSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppSorter) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (s *service) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	// filter containers for application label
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	containers, err := c.Cluster().Containers(fmt.Sprintf("labels.\"%s\"", stellar.StellarApplicationLabel))
	if err != nil {
		return nil, err
	}

	apps := map[string]*api.App{}
	for _, c := range containers {
		svc, err := s.containerToService(ctx, c)
		if err != nil {
			logrus.Error(err)
			continue
		}
		name := c.Container.Labels[stellar.StellarApplicationLabel]
		app, ok := apps[name]
		if !ok {
			app = &api.App{
				Name: name,
			}
			apps[name] = app
		}
		app.Services = append(app.Services, svc)
	}

	applications := []*api.App{}
	for _, app := range apps {
		applications = append(applications, app)
	}
	sort.Sort(AppSorter(applications))

	return &api.ListResponse{
		Applications: applications,
	}, nil
}

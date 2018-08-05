package application

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
)

func containerToService(ctx context.Context, c *clusterapi.Container) (*api.Service, error) {
	labels := []string{}
	for k, v := range c.Container.Labels {
		labels = append(labels, k+"="+v)
	}
	return &api.Service{
		Name:        c.Container.ID,
		Image:       c.Container.Image,
		Runtime:     c.Container.Runtime,
		Snapshotter: c.Container.Snapshotter,
		Labels:      labels,
	}, nil
}

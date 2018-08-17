package application

import (
	"context"

	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
)

func containerToService(ctx context.Context, c *clusterapi.Container) (*nodeapi.Service, error) {
	labels := []string{}
	for k, v := range c.Container.Labels {
		labels = append(labels, k+"="+v)
	}
	return &nodeapi.Service{
		Name:        c.Container.ID,
		Image:       c.Container.Image,
		Runtime:     c.Container.Runtime,
		Snapshotter: c.Container.Snapshotter,
		Labels:      labels,
	}, nil
}

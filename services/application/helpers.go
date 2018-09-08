package application

import (
	"context"

	"github.com/containerd/typeurl"
	"github.com/ehazlett/stellar"
	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	nodeapi "github.com/ehazlett/stellar/api/services/node/v1"
)

func (s *service) containerToService(ctx context.Context, c *clusterapi.Container) (*nodeapi.Service, error) {
	labels := []string{}
	for k, v := range c.Container.Labels {
		labels = append(labels, k+"="+v)
	}

	svc := &nodeapi.Service{
		Name:        c.Container.ID,
		Image:       c.Container.Image,
		Runtime:     c.Container.Runtime,
		Snapshotter: c.Container.Snapshotter,
		Labels:      labels,
		Node:        s.nodeName(),
	}

	// add stellar specific data
	if ext, ok := c.Container.Extensions[stellar.StellarServiceExtension]; ok {
		v, err := typeurl.UnmarshalAny(ext)
		if err != nil {
			return nil, err
		}
		s, ok := v.(*nodeapi.Service)
		if ok {
			svc.Endpoints = s.Endpoints
		}
	}

	return svc, nil
}

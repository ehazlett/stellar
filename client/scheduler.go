package client

import (
	"context"

	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
	schedulerapi "github.com/ehazlett/stellar/api/services/scheduler/v1"
)

type scheduler struct {
	client schedulerapi.SchedulerClient
}

func (s *scheduler) Schedule(service *runtimeapi.Service, nodes []*clusterapi.Node) ([]*clusterapi.Node, error) {
	ctx := context.Background()
	resp, err := s.client.Schedule(ctx, &schedulerapi.ScheduleRequest{
		Service:        service,
		AvailableNodes: nodes,
	})
	if err != nil {
		return nil, err
	}
	return resp.Nodes, nil
}

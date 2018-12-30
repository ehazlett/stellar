package scheduler

import (
	"context"

	clusterapi "github.com/ehazlett/stellar/api/services/cluster/v1"
	runtimeapi "github.com/ehazlett/stellar/api/services/runtime/v1"
	api "github.com/ehazlett/stellar/api/services/scheduler/v1"
	"github.com/sirupsen/logrus"
)

func (s *service) Schedule(ctx context.Context, req *api.ScheduleRequest) (*api.ScheduleResponse, error) {
	nodes, err := s.schedule(req.Service, req.AvailableNodes)
	if err != nil {
		return nil, err
	}
	return &api.ScheduleResponse{
		Nodes: nodes,
	}, nil
}

func (s *service) schedule(svc *runtimeapi.Service, nodes []*clusterapi.Node) ([]*clusterapi.Node, error) {
	// short-circuit to skip filtering if no preference is specified
	pref := svc.PlacementPreference
	replicas := svc.Replicas
	if replicas == 0 {
		logrus.Warn("service replicas cannot be 0; increasing to 1")
		replicas = uint64(1)
	}

	if pref == nil || len(pref.NodeIDs) == 0 && len(pref.Labels) == 0 {
		return resolveNodesForReplicas(nodes, replicas), nil
	}

	placementNodes := []*clusterapi.Node{}

	availableNodes := map[string]*clusterapi.Node{}
	for _, node := range nodes {
		availableNodes[node.ID] = node
	}

	nodeIDNodes := []*clusterapi.Node{}
	nodeLabelNodes := []*clusterapi.Node{}

	// filter node ids
	for _, id := range pref.NodeIDs {
		if node, ok := availableNodes[id]; ok {
			nodeIDNodes = append(nodeIDNodes, node)
		}
	}

	// filter node labels
	for _, node := range nodes {
		valid := false
		for k, v := range pref.Labels {
			x, ok := node.Labels[k]
			// label missing
			if !ok {
				valid = false
				break
			}
			// label value does not match and is not empty
			if x != "" && x != v {
				valid = false
				break
			}

			valid = true
		}
		if valid {
			nodeLabelNodes = append(nodeLabelNodes, node)
		}
	}

	filteredNodes := map[string]*clusterapi.Node{}
	for _, node := range nodeIDNodes {
		if _, ok := filteredNodes[node.ID]; !ok {
			filteredNodes[node.ID] = node
		}
	}
	for _, node := range nodeLabelNodes {
		if _, ok := filteredNodes[node.ID]; !ok {
			filteredNodes[node.ID] = node
		}
	}

	for _, node := range filteredNodes {
		placementNodes = append(placementNodes, node)
	}

	logrus.WithFields(logrus.Fields{
		"service":  svc.Name,
		"replicas": replicas,
	}).Debug("resolving nodes for replicas")
	return resolveNodesForReplicas(placementNodes, replicas), nil
}

func resolveNodesForReplicas(nodes []*clusterapi.Node, replicas uint64) []*clusterapi.Node {
	if len(nodes) == 0 {
		return nil
	}

	scheduledNodes := []*clusterapi.Node{}
	if uint64(len(nodes)) < replicas {
		for {
			for _, node := range nodes {
				scheduledNodes = append(scheduledNodes, node)
				if uint64(len(scheduledNodes)) == replicas {
					return scheduledNodes
				}
			}
		}
	}

	for i := uint64(0); i < replicas; i++ {
		scheduledNodes = append(scheduledNodes, nodes[i])
	}

	return scheduledNodes
}

package proxy

import (
	"context"
	"time"

	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	ptypes "github.com/gogo/protobuf/types"
)

func (s *service) Backends(ctx context.Context, req *api.BackendRequest) (*api.BackendResponse, error) {
	var backends []*api.Backend
	for _, server := range s.currentServers {
		backend := &api.Backend{
			Host: server.Host,
		}
		for _, b := range server.Backends {
			up := s.loadUpstream(b)
			backend.Upstreams = append(backend.Upstreams, up)
		}
		backends = append(backends, backend)
	}

	return &api.BackendResponse{
		Backends: backends,
	}, nil
}

func (s *service) loadUpstream(upstream *Backend) *api.Upstream {
	status := "up"
	latency, err := checkConnection(upstream.URL(), s.cfg.ProxyHealthcheckInterval)
	if err != nil {
		latency = time.Millisecond * 0
		status = err.Error()
	}

	return &api.Upstream{
		Address: upstream.Upstream,
		Latency: ptypes.DurationProto(latency),
		Status:  status,
	}

}

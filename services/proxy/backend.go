package proxy

import (
	"context"
	"time"

	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Backends(ctx context.Context, req *api.BackendRequest) (*api.BackendResponse, error) {
	var backends []*api.Backend
	logrus.Debugf("proxy: backend %+v", s.currentServers)
	for _, backend := range s.currentServers {
		upstreams := []*api.Upstream{}
		for _, srv := range backend.servers {
			status := "up"
			latency, err := checkConnection(srv, s.proxyHealthcheckInterval)
			if err != nil {
				latency = time.Millisecond * 0
				status = err.Error()
			}
			upstreams = append(upstreams, &api.Upstream{
				Address: srv.String(),
				Latency: ptypes.DurationProto(latency),
				Status:  status,
			})
		}
		backends = append(backends, &api.Backend{
			Host:      backend.host,
			Upstreams: upstreams,
		})
	}

	return &api.BackendResponse{
		Backends: backends,
	}, nil
}

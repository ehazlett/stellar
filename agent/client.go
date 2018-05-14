package agent

import (
	healthapi "github.com/ehazlett/element/api/services/health/v1"
	versionapi "github.com/ehazlett/element/api/services/version/v1"
	"google.golang.org/grpc"
)

type AgentClient struct {
	conn           *grpc.ClientConn
	VersionService versionapi.VersionClient
	HealthService  healthapi.HealthClient
}

func NewAgentClient(addr string) (*AgentClient, error) {
	c, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := &AgentClient{
		conn: c,
	}
	client.VersionService = versionapi.NewVersionClient(c)
	client.HealthService = healthapi.NewHealthClient(c)

	return client, nil
}

func (a *AgentClient) Close() error {
	return a.conn.Close()
}

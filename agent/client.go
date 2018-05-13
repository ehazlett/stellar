package agent

import (
	versionapi "github.com/ehazlett/element/api/services/version/v1"
	"google.golang.org/grpc"
)

type AgentClient struct {
	VersionService versionapi.VersionClient
}

func NewAgentClient(addr string) (*AgentClient, error) {
	c, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := &AgentClient{}
	client.VersionService = versionapi.NewVersionClient(c)

	return client, nil
}

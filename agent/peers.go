package agent

import (
	"encoding/json"
	"fmt"
	"time"
)

// PeerAgent is the peer information for an agent in the cluster including name and GRPC address
type PeerAgent struct {
	Name    string
	Addr    string
	Updated time.Time
}

// Peers returns all known peers in the cluster
func (a *Agent) Peers() ([]*PeerAgent, error) {
	self := a.members.LocalNode()
	var (
		peerAgents map[string]*PeerAgent
		peers      []*PeerAgent
	)
	if err := json.Unmarshal(self.Meta, &peerAgents); err != nil {
		return nil, err
	}

	for _, p := range peerAgents {
		peers = append(peers, p)
	}

	return peers, nil
}

// LocalNode returns local node peer info
func (a *Agent) LocalNode() (*PeerAgent, error) {
	return &PeerAgent{
		Name:    a.config.NodeName,
		Addr:    fmt.Sprintf("%s:%d", a.config.AgentAddr, a.config.AgentPort),
		Updated: time.Now(),
	}, nil
}

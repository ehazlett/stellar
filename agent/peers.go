package agent

import (
	"encoding/json"
	"time"
)

type PeerAgent struct {
	Name    string
	Addr    string
	Updated time.Time
}

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

package element

import "github.com/gogo/protobuf/proto"

// Peers returns all known peers in the cluster
func (a *Agent) Peers() ([]*Peer, error) {
	self := a.members.LocalNode()
	var state State
	if err := proto.Unmarshal(self.Meta, &state); err != nil {
		return nil, err
	}
	var peers []*Peer
	for _, p := range state.Peers {
		peers = append(peers, p)
	}
	return peers, nil
}

// Self returns the local peer information
func (a *Agent) Self() *Peer {
	return a.state.Self
}

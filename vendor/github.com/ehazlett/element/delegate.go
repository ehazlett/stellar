package element

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
)

// NodeMeta returns local node meta information
func (a *Agent) NodeMeta(limit int) []byte {
	data, err := proto.Marshal(a.state)
	if err != nil {
		logrus.Errorf("error serializing node meta: %s", err)
	}
	return data
}

// NotifyMsg is used for handling cluster messages
func (a *Agent) NotifyMsg(buf []byte) {
	// this can be used to receive messages sent (i.e. SendReliable)
}

// GetBroadcasts is called when user messages can be broadcast
func (a *Agent) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

// LocalState is the local cluster agent state
func (a *Agent) LocalState(join bool) []byte {
	data, err := proto.Marshal(a.state)
	if err != nil {
		logrus.Errorf("error serializing local state: %s", err)
	}
	return data
}

// MergeRemoteState is used to store remote peer information
func (a *Agent) MergeRemoteState(buf []byte, join bool) {
	var state State
	if err := proto.Unmarshal(buf, &state); err != nil {
		logrus.Errorf("error parsing remote agent state: %s", err)
		return
	}
	a.state.Updated = time.Now()
	a.state.Peers[state.Self.ID] = state.Self
	// notify update
	a.peerUpdateChan <- true
}

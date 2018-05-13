package agent

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

type agentDelegate struct {
	Name       string
	Addr       string
	Peers      map[string]*PeerAgent
	updateChan chan bool
}

func NewAgentDelegate(name, addr string, ch chan bool) *agentDelegate {
	agent := &agentDelegate{
		Name:       name,
		Addr:       addr,
		Peers:      make(map[string]*PeerAgent),
		updateChan: ch,
	}

	t := time.NewTicker(nodeHeartbeatInterval)
	go func() {
		for range t.C {
			// reconcile
			agent.reconcile()
		}
	}()

	return agent
}

func (d *agentDelegate) reconcile() {
	changed := false
	for name, peer := range d.Peers {
		if time.Now().After(peer.Updated.Add(nodeReconcileTimeout)) {
			logrus.Debugf("peer timeout; removing %s", name)
			delete(d.Peers, name)
			changed = true
		}
	}
	if changed {
		d.updateChan <- true
	}
}

func (d *agentDelegate) NodeMeta(limit int) []byte {
	data, err := json.Marshal(d.Peers)
	if err != nil {
		logrus.Errorf("error serializing node meta: %s", err)
	}
	return data
}

func (d *agentDelegate) NotifyMsg(buf []byte) {
	// this can be used to receive messages sent (i.e. SendReliable)
}

// GetBroadcasts is called when user messages can be broadcast
func (d *agentDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

func (d *agentDelegate) LocalState(join bool) []byte {
	data, err := json.Marshal(d)
	if err != nil {
		logrus.Errorf("error serializing local state: %s", err)
	}
	return []byte(data)
}

func (d *agentDelegate) MergeRemoteState(buf []byte, join bool) {
	var remoteAgent *agentDelegate
	if err := json.Unmarshal(buf, &remoteAgent); err != nil {
		logrus.Errorf("error parsing remote agent state: %s", err)
		return
	}
	d.Peers[remoteAgent.Name] = &PeerAgent{
		Name:    remoteAgent.Name,
		Addr:    remoteAgent.Addr,
		Updated: time.Now(),
	}
	// notify update
	d.updateChan <- true
}

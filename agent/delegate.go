package agent

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

type agentDelegate struct {
	Name          string
	Addr          string
	Peers         map[string]*PeerAgent
	updateChan    chan bool
	nodeEventChan chan *NodeEvent
}

func NewAgentDelegate(name, addr string, updateCh chan bool, nodeEventCh chan *NodeEvent) *agentDelegate {
	agent := &agentDelegate{
		Name:          name,
		Addr:          addr,
		Peers:         make(map[string]*PeerAgent),
		updateChan:    updateCh,
		nodeEventChan: nodeEventCh,
	}

	// event handler
	go func() {
		for {
			select {
			case evt := <-nodeEventCh:
				switch evt.EventType {
				case NodeJoin:
				case NodeUpdate:
				case NodeLeave:
					agent.removeNode(evt.Node.Name)
				}
			}
		}
	}()

	return agent
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

func (d *agentDelegate) removeNode(name string) {
	if _, exists := d.Peers[name]; exists {
		delete(d.Peers, name)
		// notify update
		d.updateChan <- true
	}
}

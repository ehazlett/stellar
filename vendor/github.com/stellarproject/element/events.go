package element

import (
	"github.com/hashicorp/memberlist"
)

// NodeEventType is the type of node event
type NodeEventType string

const (
	// NodeJoin is the event fired upon a node joining the cluster
	NodeJoin NodeEventType = "join"
	// NodeLeave is the event fired upon a node leaving the cluster
	NodeLeave NodeEventType = "leave"
	// NodeUpdate is the event fired upon a node updating in the cluster
	NodeUpdate NodeEventType = "update"
)

// NodeEvent stores the event type and node information
type NodeEvent struct {
	// EventType is the type of event fired
	Type NodeEventType
	// Node is the internal cluster node
	Node *memberlist.Node
}

// NotifyJoin notifies when a node joins the cluster
func (a *Agent) NotifyJoin(n *memberlist.Node) {
	a.send(&NodeEvent{
		Type: NodeJoin,
		Node: n,
	})
}

// NotifyLeave notifies when a node leaves the cluster
func (a *Agent) NotifyLeave(n *memberlist.Node) {
	delete(a.state.Peers, n.Name)
	a.peerUpdateChan <- true
	a.send(&NodeEvent{
		Type: NodeLeave,
		Node: n,
	})
}

// NotifyUpdate notifies when a node is updated in the cluster
func (a *Agent) NotifyUpdate(n *memberlist.Node) {
	a.send(&NodeEvent{
		Type: NodeUpdate,
		Node: n,
	})
}

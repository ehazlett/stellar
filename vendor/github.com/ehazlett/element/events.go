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
	EventType NodeEventType
	// Node is the internal cluster node
	Node *memberlist.Node
}

// EventHandler is used for event handling
type EventHandler struct {
	ch chan *NodeEvent
}

// NewEventHandler returns an EventHandler that is used to perform actions for the specified event
func NewEventHandler(ch chan *NodeEvent) *EventHandler {
	return &EventHandler{
		ch: ch,
	}
}

// NotifyJoin notifies when a node joins the cluster
func (h *EventHandler) NotifyJoin(n *memberlist.Node) {
	go h.notify(NodeJoin, n)
}

// NotifyLeave notifies when a node leaves the cluster
func (h *EventHandler) NotifyLeave(n *memberlist.Node) {
	go h.notify(NodeLeave, n)
}

// NotifyUpdate notifies when a node is updated in the cluster
func (h *EventHandler) NotifyUpdate(n *memberlist.Node) {
	go h.notify(NodeUpdate, n)
}

func (h *EventHandler) notify(t NodeEventType, n *memberlist.Node) {
	// TODO: use context WithTimeout to enable cancel
	h.ch <- &NodeEvent{
		EventType: t,
		Node:      n,
	}
}

package agent

import (
	"github.com/hashicorp/memberlist"
)

type NodeEventType string

const (
	NodeJoin   NodeEventType = "join"
	NodeLeave  NodeEventType = "leave"
	NodeUpdate NodeEventType = "update"
)

type NodeEvent struct {
	EventType NodeEventType
	Node      *memberlist.Node
}

type EventHandler struct {
	ch chan *NodeEvent
}

func NewEventHandler(ch chan *NodeEvent) *EventHandler {
	return &EventHandler{
		ch: ch,
	}
}

func (h *EventHandler) NotifyJoin(n *memberlist.Node) {
	go h.notify(NodeJoin, n)
}

func (h *EventHandler) NotifyLeave(n *memberlist.Node) {
	go h.notify(NodeLeave, n)
}

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

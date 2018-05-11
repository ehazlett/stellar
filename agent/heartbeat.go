package agent

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func (a *Agent) heartbeat() {
	peers := []string{}
	self := a.members.LocalNode()
	for _, node := range a.members.Members() {
		// ignore self
		if node.Name == self.Name {
			continue
		}
		peers = append(peers, fmt.Sprintf("%s (%s)", node.Name, node.Addr))
	}
	if len(peers) > 0 {
		logrus.Debugf("peers: %s", strings.Join(peers, ", "))
	}
}

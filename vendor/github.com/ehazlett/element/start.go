package element

import (
	"github.com/sirupsen/logrus"
)

// Start handles cluster events
func (a *Agent) Start() error {
	go func() {
		for range a.peerUpdateChan {
			if err := a.members.UpdateNode(nodeUpdateTimeout); err != nil {
				logrus.Errorf("error updating node metadata: %s", err)
			}
		}
	}()
	if len(a.config.Peers) > 0 {
		if _, err := a.members.Join(a.config.Peers); err != nil {
			return err
		}
	}
	return nil
}

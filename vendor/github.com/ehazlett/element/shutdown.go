package element

import (
	"github.com/sirupsen/logrus"
)

// Shutdown causes the local node to leave the cluster and perform a clean shutdown
func (a *Agent) Shutdown() error {
	logrus.Debug("agent leave")
	if err := a.members.Leave(nodeUpdateTimeout); err != nil {
		return err
	}

	logrus.Debug("agent shutdown")
	if err := a.members.Shutdown(); err != nil {
		return err
	}

	return nil
}

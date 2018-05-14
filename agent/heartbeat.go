package agent

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (a *Agent) heartbeat() {
	peers, err := a.Peers()
	if err != nil {
		logrus.Errorf("error getting peers: %s", err)
		return
	}

	for _, peer := range peers {
		ac, err := NewAgentClient(peer.Addr)
		if err != nil {
			logrus.Errorf("error communicating with peer: %s", err)
			return
		}
		defer ac.Close()

		health, err := ac.HealthService.Health(context.Background(), nil)
		if err != nil {
			logrus.Errorf("error communicating with peer: %s", err)
			return
		}

		logrus.WithFields(logrus.Fields{
			"peer":       peer.Name,
			"os_name":    health.OsName,
			"os_version": health.OsVersion,
			"uptime":     health.Uptime,
			"containers": health.Containers,
			"images":     health.Images,
		}).Debug("peer health")

		//v, err := ac.VersionService.Version(context.Background(), nil)
		//if err != nil {
		//	logrus.Error(err)
		//}
	}
}

package agent

import (
	"context"
	"strings"
	"time"

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
			"peer":         peer.Name,
			"os_name":      health.OSName,
			"os_version":   health.OSVersion,
			"uptime":       health.Uptime,
			"cpus":         health.Cpus,
			"memory_total": health.MemoryTotal,
			"memory_free":  health.MemoryFree,
			"memory_used":  health.MemoryUsed,
		}).Debug("peer health")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		resp, err := ac.NodeService.Containers(ctx, nil)
		if err != nil {
			logrus.Errorf("error getting containers: %s", err)
			return
		}

		ids := []string{}
		for _, c := range resp.Containers {
			ids = append(ids, c.ID)
		}

		logrus.WithFields(logrus.Fields{
			"peer":       peer.Name,
			"containers": strings.Join(ids, ", "),
		}).Debug("containers")
	}
}

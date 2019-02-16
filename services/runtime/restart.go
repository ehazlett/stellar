package runtime

import (
	"context"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/ehazlett/stellar"
	"github.com/sirupsen/logrus"
)

func (s *service) restartMonitor() {
	t := time.NewTicker(s.restartInterval)
	for range t.C {
		c, err := s.containerd()
		if err != nil {
			logrus.WithError(err).Error("unable to connect to containerd")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		containers, err := c.Containers(ctx)
		if err != nil {
			logrus.WithError(err).Error("unable to list containers")
			c.Close()
			continue
		}

		for _, container := range containers {
			labels, err := container.Labels(ctx)
			if err != nil {
				logrus.WithError(err).Error("error getting container labels")
				continue
			}
			if _, ok := labels[stellar.StellarRestartLabel]; ok {
				if err := s.ensureRunning(container); err != nil {
					logrus.WithError(err).Error("error ensuring container is running")
					continue
				}
			}
		}

		c.Close()
	}
}

func (s *service) ensureRunning(container containerd.Container) error {
	c, err := s.containerd()
	if err != nil {
		return err
	}
	defer c.Close()

	logrus.WithFields(logrus.Fields{
		"container": container.ID(),
	}).Debug("checking restart for container")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	task, err := container.Task(ctx, nil)
	if err != nil {
		if !errdefs.IsNotFound(err) {
			return err
		}

		// task not found; restart
		logrus.WithFields(logrus.Fields{
			"container": container.ID(),
		}).Debug("container task not found; restarting")
		if err := s.startTask(ctx, container); err != nil {
			return err
		}

		return nil
	}

	st, err := task.Status(ctx)
	if err != nil {
		return err
	}

	status := st.Status

	switch status {
	case containerd.Running:
		// task is running
		return nil
	case containerd.Created, containerd.Stopped, containerd.Unknown:
		// remove existing task and restart
		if err := s.killTask(ctx, task); err != nil {
			logrus.WithFields(logrus.Fields{
				"container": container.ID(),
			}).Warnf("error deleting existing task: %s", err)
		}
		// restart
		if err := s.startTask(ctx, container); err != nil {
			return err
		}
	case containerd.Paused, containerd.Pausing:
		logrus.WithFields(logrus.Fields{
			"container": container.ID(),
			"status":    status,
		}).Info("container is paused; not restarting")
	default:
		logrus.WithFields(logrus.Fields{
			"container": container.ID(),
			"status":    status,
		}).Warn("unknown container status; not restarting")
	}

	return nil
}

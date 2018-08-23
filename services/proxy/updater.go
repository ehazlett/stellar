package proxy

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func (s *service) updater() {
	for u := range s.updateCh {
		logrus.Debugf("proxy: action=%s server %s", u.action, u.backend.host)

		host := fmt.Sprintf(`Host("%s")`, u.backend.host)
		switch u.action {
		case updateActionAdd:
			logrus.Debugf("proxy: adding server %s", u.backend.host)
			for _, server := range u.backend.servers {
				if err := u.backend.lb.UpsertServer(server); err != nil {
					s.errCh <- err
				}
			}
			if err := s.mux.Handle(host, u.backend.lb); err != nil {
				s.errCh <- err
			}
		case updateActionUpdate:
			logrus.Debugf("proxy: updating server %s", u.backend.host)
			for _, server := range u.backend.servers {
				if err := u.backend.lb.UpsertServer(server); err != nil {
					s.errCh <- err
				}
			}
		case updateActionRemove:
			logrus.Debugf("proxy: removing server %s", u.backend.host)
			for _, server := range u.backend.servers {
				if err := u.backend.lb.RemoveServer(server); err != nil {
					s.errCh <- err
				}
			}
			if err := s.mux.Remove(host); err != nil {
				s.errCh <- err
			}
		}
	}
}

package proxy

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func (s *service) healthcheck() {
	logrus.Debugf("proxy: starting healthcheck interval=%s", s.proxyHealthcheckInterval)
	t := time.NewTicker(s.proxyHealthcheckInterval)
	checking := false
	for range t.C {
		// only allow a single set of checks
		if checking {
			continue
		}
		checking = true
		wg := &sync.WaitGroup{}
		// check backends
		for _, backend := range s.currentServers {
			wg.Add(1)
			go s.checkBackend(backend, wg)
		}
		wg.Wait()
		checking = false
	}
}

func (s *service) checkBackend(backend *backend, wg *sync.WaitGroup) {
	defer wg.Done()

	//toRemove := []int{}
	for _, server := range backend.servers {
		if _, err := checkConnection(server, s.proxyHealthcheckInterval); err != nil {
			// TODO: remove server
			logrus.Warnf("proxy: error accessing upstream %s for host %s; removing", server.Host, backend.host)
			if err := backend.lb.RemoveServer(server); err != nil {
				s.errCh <- err
			}
			//toRemove = append(toRemove, i)
		}
	}

	//for i := range toRemove {
	//	backend.servers = append(backend.servers[:i], backend.servers[i+1:]...)
	//}
}

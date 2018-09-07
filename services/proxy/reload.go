package proxy

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	nameserverapi "github.com/ehazlett/stellar/api/services/nameserver/v1"
	api "github.com/ehazlett/stellar/api/services/proxy/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Reload(ctx context.Context, req *api.ReloadRequest) (*ptypes.Empty, error) {
	if err := s.reload(); err != nil {
		return empty, err
	}
	return empty, nil
}

func (s *service) reload() error {
	c, err := s.client()
	if err != nil {
		return err
	}
	defer c.Close()

	apps, err := c.Application().List()
	if err != nil {
		return err
	}

	next := map[string]*backend{}
	appServers := map[string][]*url.URL{}

	for _, app := range apps {
		for _, svc := range app.Services {
			if len(svc.Endpoints) == 0 {
				continue
			}
			// parse url and notify update
			for _, ep := range svc.Endpoints {
				if strings.ToLower(ep.Protocol.String()) != "http" {
					logrus.Warnf("proxy: unsupported protocol %s for endpoint %s", ep.Protocol, ep.Service)
					continue
				}
				// lookup the ip for faster resolves
				records, err := c.Nameserver().Lookup(svc.Name)
				if err != nil {
					logrus.Warnf("proxy: unable to lookup service address: %s", err)
					continue
				}
				servers := []*url.URL{}
				for _, record := range records {
					// only use A records
					if record.Type != nameserverapi.RecordType_A {
						continue
					}
					servers = append(servers, &url.URL{
						Scheme: "http",
						Host:   fmt.Sprintf("%s:%d", record.Value, ep.Port),
						Path:   "/",
					})
				}
				appServers[ep.Host] = append(appServers[ep.Host], servers...)
			}
		}
	}

	// trigger updates
	for host, servers := range appServers {
		serverID, err := generateID(host)
		if err != nil {
			logrus.Errorf("proxy: error generating server id: %s", err)
			continue
		}

		b, exists := s.currentServers[serverID]
		if !exists {
			lb, err := newLB()
			if err != nil {
				logrus.Errorf("proxy: error setting up backend: %s", err)
				continue
			}
			b = &backend{
				host:    host,
				lb:      lb,
				servers: servers,
			}
		}

		// schedule update
		update := &proxyUpdate{
			backend: b,
		}
		update.action = s.getUpdateAction(serverID)
		updates = append(updates, update)
		next[serverID] = b
	}

	// notify updates
	for _, update := range updates {
		logrus.Debugf("proxy update: host=%s servers=%+v", update.backend.host, update.backend.servers)
		s.updateCh <- update
	}
	// update current servers
	s.currentServers = next
	// prune
	s.pruneServers(next)

	return nil
}

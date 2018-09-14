package proxy

import (
	"fmt"
	"strings"

	blackbirdapi "github.com/ehazlett/blackbird/api/v1"
	nameserverapi "github.com/ehazlett/stellar/api/services/nameserver/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/sirupsen/logrus"
)

type datastore struct {
	client *client.Client
}

func newDatastore(client *client.Client) (*datastore, error) {
	if err := client.Datastore().CreateBucket(dsProxyBucketName); err != nil {
		return nil, err
	}
	return &datastore{
		client: client,
	}, nil
}

func (d *datastore) Name() string {
	return "stellar"
}

func (d *datastore) Add(host string, srv *blackbirdapi.Server) error {
	return nil
}

func (d *datastore) Remove(host string) error {
	return nil
}

func (d *datastore) Servers() ([]*blackbirdapi.Server, error) {
	apps, err := d.client.Application().List()
	if err != nil {
		return nil, err
	}

	servers := []*blackbirdapi.Server{}

	hostServers := map[string]*blackbirdapi.Server{}
	hostUpstreams := map[string][]string{}
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
				records, err := d.client.Nameserver().Lookup(svc.Name)
				if err != nil {
					logrus.Warnf("proxy: unable to lookup service address: %s", err)
					continue
				}
				v, exists := hostServers[ep.Host]
				if !exists {
					v = &blackbirdapi.Server{
						Host:   ep.Host,
						Path:   "/",
						TLS:    ep.TLS,
						Policy: blackbirdapi.Policy_RANDOM,
						Preset: "transparent",
					}
					hostServers[ep.Host] = v
				}
				// check for conflicting values in host
				if ep.TLS != v.TLS {
					logrus.Warnf("conflicting TLS setting for %s", ep.Host)
				}

				// build upstreams
				upstreams := []string{}
				for _, record := range records {
					// only use A records
					if record.Type != nameserverapi.RecordType_A {
						continue
					}
					// TODO: support TLS on the upstream
					upstreams = append(upstreams, fmt.Sprintf("http://%s:%d", record.Value, ep.Port))
				}
				if _, ok := hostUpstreams[ep.Host]; !ok {
					hostUpstreams[ep.Host] = []string{}
				}
				hostUpstreams[ep.Host] = append(hostUpstreams[ep.Host], upstreams...)
			}
		}
	}

	for host, srv := range hostServers {
		up, ok := hostUpstreams[host]
		if !ok {
			logrus.Warnf("no upstreams found for %s", host)
		}
		srv.Upstreams = up
		servers = append(servers, srv)
	}
	return servers, nil
}

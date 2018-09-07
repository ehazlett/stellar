package proxy

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	nameserverapi "github.com/ehazlett/stellar/api/services/nameserver/v1"
	"github.com/sirupsen/logrus"
)

func (s *service) getApplicationEndpoints() (map[string][]*url.URL, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	apps, err := c.Application().List()
	if err != nil {
		return nil, err
	}

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

	return appServers, nil
}

func generateID(v interface{}, extra ...interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write(data)
	for _, x := range extra {
		d, err := json.Marshal(x)
		if err != nil {
			return "", err

		}
		h.Write(d)
	}
	r := hex.EncodeToString(h.Sum(nil))[:24]
	return r, nil
}

func checkConnection(endpoint *url.URL, timeout time.Duration) (time.Duration, error) {
	zero := time.Millisecond * 0
	switch endpoint.Scheme {
	case "tcp":
		c, err := net.DialTimeout("tcp", endpoint.Host, timeout)
		if err != nil {
			return zero, err
		}
		start := time.Now()
		if _, err := c.Read([]byte{}); err != nil {
			return zero, err
		}

		latency := time.Now().Sub(start)
		return latency, nil
	case "http", "https":
		c := &http.Client{
			Timeout: timeout,
		}
		start := time.Now()
		if _, err := c.Get(endpoint.Scheme + "://" + endpoint.Host); err != nil {
			return zero, err
		}
		latency := time.Now().Sub(start)
		return latency, nil
	}

	return zero, nil
}

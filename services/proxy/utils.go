package proxy

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

//func (s *service) getServers() ([]*blackbirdapi.Server, error) {
//	c, err := s.client()
//	if err != nil {
//		return nil, err
//	}
//	defer c.Close()
//
//	apps, err := c.Application().List()
//	if err != nil {
//		return nil, err
//	}
//
//	servers := []*blackbirdapi.Server{}
//
//	hostServers := map[string]*blackbirdapi.Server{}
//	hostUpstreams := map[string][]string{}
//	for _, app := range apps {
//		for _, svc := range app.Services {
//			if len(svc.Endpoints) == 0 {
//				continue
//			}
//			// parse url and notify update
//			for _, ep := range svc.Endpoints {
//				if strings.ToLower(ep.Protocol.String()) != "http" {
//					logrus.Warnf("proxy: unsupported protocol %s for endpoint %s", ep.Protocol, ep.Service)
//					continue
//				}
//				// lookup the ip for faster resolves
//				records, err := c.Nameserver().Lookup(svc.Name)
//				if err != nil {
//					logrus.Warnf("proxy: unable to lookup service address: %s", err)
//					continue
//				}
//				v, exists := hostServers[ep.Host]
//				if !exists {
//					v = &blackbirdapi.Server{
//						Host:   ep.Host,
//						Path:   "/",
//						TLS:    ep.TLS,
//						Policy: blackbirdapi.Policy_RANDOM,
//						Preset: "transparent",
//					}
//					hostServers[ep.Host] = v
//				}
//				// check for conflicting values in host
//				if ep.TLS != v.TLS {
//					logrus.Warnf("conflicting TLS setting for %s", ep.Host)
//				}
//
//				// build upstreams
//				upstreams := []string{}
//				for _, record := range records {
//					// only use A records
//					if record.Type != nameserverapi.RecordType_A {
//						continue
//					}
//					// TODO: support TLS on the upstream
//					upstreams = append(upstreams, fmt.Sprintf("http://%s:%d", record.Value, ep.Port))
//				}
//				if _, ok := hostUpstreams[ep.Host]; !ok {
//					hostUpstreams[ep.Host] = []string{}
//				}
//				hostUpstreams[ep.Host] = append(hostUpstreams[ep.Host], upstreams...)
//			}
//		}
//	}
//
//	for host, srv := range hostServers {
//		up, ok := hostUpstreams[host]
//		if !ok {
//			logrus.Warnf("no upstreams found for %s", host)
//		}
//		srv.Upstreams = up
//		servers = append(servers, srv)
//	}
//
//	return servers, nil
//}
//
//func generateID(v interface{}, extra ...interface{}) (string, error) {
//	data, err := json.Marshal(v)
//	if err != nil {
//		return "", err
//	}
//	h := sha1.New()
//	h.Write(data)
//	for _, x := range extra {
//		d, err := json.Marshal(x)
//		if err != nil {
//			return "", err
//
//		}
//		h.Write(d)
//	}
//	r := hex.EncodeToString(h.Sum(nil))[:24]
//	return r, nil
//}
//
func checkConnection(upstream string, timeout time.Duration) (time.Duration, error) {
	zero := time.Millisecond * 0
	endpoint, err := url.Parse(upstream)
	if err != nil {
		return zero, err
	}

	switch endpoint.Scheme {
	case "tcp":
		c, err := net.DialTimeout("tcp", endpoint.Host, timeout)
		if err != nil {
			return zero, err
		}
		defer c.Close()
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

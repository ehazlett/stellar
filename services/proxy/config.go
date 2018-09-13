package proxy

import (
	"bytes"
	"html/template"
	"net/url"
)

const (
	configTemplate = `# stellar proxy config
*:{{ $.HTTPPort }} {
    status 200 /healthz
}
{{ range $server := .Servers }}
{{ if $server.TLS }}https{{ else }}http{{ end }}://{{ $server.Host }}:{{ if $server.TLS }}{{ $.HTTPSPort }}{{ else }}{{ $.HTTPPort }}{{ end }} {
    proxy {{ $server.Path }} {
	transparent
	policy random
	try_duration 1s
	{{ range $backend := $server.Backends }}upstream {{ .Upstream }}
	{{ end }}
    }
    timeouts none
}
{{ end }}
`
)

type Backend struct {
	Scheme   string
	Host     string
	Port     string
	Path     string
	Upstream string
}

func (b *Backend) URL() *url.URL {
	return &url.URL{
		Scheme: b.Scheme,
		Host:   b.Upstream,
		Path:   b.Path,
	}
}

type Server struct {
	Host     string
	Path     string
	TLS      bool
	Backends []*Backend
}

type Config struct {
	HTTPPort  int
	HTTPSPort int
	Servers   []*Server
}

func (s *service) generateConfig() ([]byte, error) {
	t := template.New("proxy")
	tmpl, err := t.Parse(configTemplate)
	if err != nil {
		return nil, err
	}

	config := &Config{
		HTTPPort:  s.cfg.ProxyHTTPPort,
		HTTPSPort: s.cfg.ProxyHTTPSPort,
	}
	// backends is a map[host]map[path][]Backend to the backend upstreams
	srvs := map[string]*Server{}
	apps, err := s.getApplicationEndpoints()
	if err != nil {
		return nil, err
	}
	for host, app := range apps {
		for _, ep := range app {
			srv, exists := srvs[host]
			if !exists {
				srv = &Server{
					Host:     host,
					Path:     ep.url.Path,
					TLS:      ep.tls,
					Backends: []*Backend{},
				}
				srvs[host] = srv
			}
			srv.Backends = append(srv.Backends, &Backend{
				Scheme:   ep.url.Scheme,
				Host:     host,
				Port:     ep.url.Port(),
				Path:     ep.url.Path,
				Upstream: ep.url.Host,
			})
		}
	}
	servers := []*Server{}
	for _, srv := range srvs {
		servers = append(servers, srv)
	}
	config.Servers = servers

	var c bytes.Buffer
	if err := tmpl.Execute(&c, config); err != nil {
		return nil, err
	}

	s.currentServers = servers

	return c.Bytes(), nil
}

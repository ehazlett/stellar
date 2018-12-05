package server

import (
	"bytes"
	"context"
	"strings"
	"text/template"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	api "github.com/stellarproject/radiant/api/v1"
)

const (
	configTemplate = `# radiant proxy config
*:{{ $.HTTPPort }} {
    status 200 /healthz
}
{{ range $server := .Servers }}
{{ if $server.TLS }}https{{ else }}http{{ end }}://{{ $server.Host }}:{{ if $server.TLS }}{{ $.HTTPSPort }}{{ else }}{{ $.HTTPPort }}{{ end }} {
    proxy {{ $server.Path }} { {{ if ne $server.Preset "" }}
	{{ $server.Preset }}{{ end }}
	policy {{policyname $server.Policy }} {{ if $t := $server.ProxyTryDuration }}
	try_duration {{duration $t }}{{ end }} {{ if $t := $server.ProxyFailTimeout }}
	fail_timeout {{duration $t }}{{ end }} {{if $server.ProxyUpstreamHeaders}}
	{{ range $k, $v := $server.ProxyUpstreamHeaders }}header_upstream {{ $k }} "{{ $v }}"
	{{ end }}{{ end }}
	{{ range $upstream := $server.Upstreams }}upstream {{ $upstream }}
	{{ end }}
    }
    {{ with $t := $server.Timeouts }}{{ if $t }}timeouts {{duration $t }} {{ end }} {{ end }}
}{{ end }}
`
)

type proxyConfig struct {
	HTTPPort  int
	HTTPSPort int
	Servers   []*api.Server
}

func (s *Server) Config(ctx context.Context, req *api.ConfigRequest) (*api.ConfigResponse, error) {
	cf := s.instance.Caddyfile()
	return &api.ConfigResponse{
		Data: cf.Body(),
	}, nil
}

func policyName(v api.Policy) string {
	return strings.ToLower(v.String())
}

func duration(v *types.Duration) time.Duration {
	d, err := types.DurationFromProto(v)
	if err != nil {
		logrus.Error(errors.Wrap(err, "error converting proto duration"))
	}
	return d
}

func (s *Server) generateServerConfig() ([]byte, error) {
	servers, err := s.datastore.Servers()
	if err != nil {
		return nil, err
	}
	srvs := []*api.Server{}
	for _, srv := range servers {
		srvs = append(srvs, srv)
	}

	b, err := generateConfig(s.config.HTTPPort, s.config.HTTPSPort, srvs)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateConfig(httpPort, httpsPort int, servers []*api.Server) ([]byte, error) {
	t := template.New("radiant").Funcs(template.FuncMap{
		"policyname": policyName,
		"duration":   duration,
	})
	tmpl, err := t.Parse(configTemplate)
	if err != nil {
		return nil, err
	}

	config := &proxyConfig{
		HTTPPort:  httpPort,
		HTTPSPort: httpsPort,
		Servers:   servers,
	}

	var c bytes.Buffer
	if err := tmpl.Execute(&c, config); err != nil {
		return nil, err
	}

	return c.Bytes(), nil
}

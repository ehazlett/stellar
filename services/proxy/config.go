package proxy

import (
	"bytes"
	"html/template"
)

const (
	configTemplate = `# stellar proxy config
:{{.HTTPPort}}
`
)

type Config struct {
	HTTPPort  int
	HTTPSPort int
}

func (s *service) generateConfig() ([]byte, error) {
	t := template.New("proxy")
	tmpl, err := t.Parse(configTemplate)
	if err != nil {
		return nil, err
	}

	var c bytes.Buffer
	if err := tmpl.Execute(&c, Config{
		HTTPPort:  s.cfg.ProxyHTTPPort,
		HTTPSPort: s.cfg.ProxyHTTPSPort,
	}); err != nil {
		return nil, err
	}

	return c.Bytes(), nil
}

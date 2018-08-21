package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

type Config struct {
	HTTPPort int
	Client   *client.Client
}

type Proxy struct {
	config   *Config
	lb       *roundrobin.RoundRobin
	errCh    chan error
	updateCh chan *url.URL
}

func NewProxy(cfg *Config) (*Proxy, error) {
	fwd, err := forward.New()
	if err != nil {
		return nil, err
	}
	lb, err := roundrobin.New(fwd)
	if err != nil {
		return nil, err
	}

	errCh := make(chan error)
	go func() {
		for {
			errCh <- err
			logrus.Errorf("proxy: %s", err)
		}
	}()

	return &Proxy{
		config:   cfg,
		lb:       lb,
		errCh:    errCh,
		updateCh: make(chan *url.URL),
	}, nil
}

func (p *Proxy) updater() {
	for u := range p.updateCh {
		logrus.Debugf("proxy: adding server %s", u)
		// TODO: handle remove / updates
		_ = p.lb.UpsertServer(u)
	}
}

func (p *Proxy) Reload(apps []*applicationapi.App) error {
	// TODO: get current app list and trigger update
	for _, app := range apps {
		for _, svc := range app.Services {
			if len(svc.Endpoints) == 0 {
				continue
			}

			// parse url and notify update
			for _, ep := range svc.Endpoints {
				logrus.Debugf("proxy: checking endpoint %s for app %s", ep.Service, app.Name)
				if strings.ToLower(ep.Protocol.String()) != "http" {
					logrus.Warnf("proxy: unsupported protocol %s for endpoint %s", ep.Protocol, ep.Service)
					continue
				}
				p.updateCh <- &url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s:%d", svc.Name, ep.Port),
					Path:   "/",
				}
			}
		}
	}

	return nil
}

func (p *Proxy) Run() error {
	go p.updater()

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.config.HTTPPort),
		Handler: p.lb,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			logrus.Error(errors.Wrap(err, "proxy"))
		}
	}()

	return nil
}

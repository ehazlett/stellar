package proxy

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	applicationapi "github.com/ehazlett/stellar/api/services/application/v1"
	"github.com/ehazlett/stellar/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/route"
)

type Config struct {
	HTTPPort int
	Client   *client.Client
}

type Proxy struct {
	config         *Config
	errCh          chan error
	updateCh       chan *proxyUpdate
	currentServers map[string]*backend
	mux            *route.Mux
}

type backend struct {
	host    string
	lb      *roundrobin.RoundRobin
	servers []*url.URL
}

type updateAction string

const (
	updateActionAdd    updateAction = "add"
	updateActionUpdate updateAction = "update"
	updateActionRemove updateAction = "remove"
)

type proxyUpdate struct {
	action  updateAction
	backend *backend
}

func NewProxy(cfg *Config) (*Proxy, error) {
	errCh := make(chan error)
	go func() {
		for {
			err := <-errCh
			logrus.Errorf("proxy: %s", err)
		}
	}()

	return &Proxy{
		config:   cfg,
		errCh:    errCh,
		updateCh: make(chan *proxyUpdate),
		mux:      route.NewMux(),
	}, nil
}

func newLB() (*roundrobin.RoundRobin, error) {
	// TODO: log separately?
	l := logrus.New()
	l.Out = ioutil.Discard
	fwd, err := forward.New(forward.Logger(l))
	if err != nil {
		return nil, err
	}
	lb, err := roundrobin.New(fwd, roundrobin.RoundRobinLogger(l))
	if err != nil {
		return nil, err
	}

	return lb, nil
}

func (p *Proxy) updater() {
	for u := range p.updateCh {
		logrus.Debugf("proxy: action=%s server %s", u.action, u.backend.host)

		host := fmt.Sprintf(`Host("%s")`, u.backend.host)
		switch u.action {
		case updateActionAdd:
			logrus.Debugf("proxy: adding server %s", u.backend.host)
			for _, server := range u.backend.servers {
				if err := u.backend.lb.UpsertServer(server); err != nil {
					p.errCh <- err
				}
			}
			if err := p.mux.Handle(host, u.backend.lb); err != nil {
				p.errCh <- err
			}
		case updateActionUpdate:
			logrus.Debugf("proxy: updating server %s", u.backend.host)
			for _, server := range u.backend.servers {
				if err := u.backend.lb.UpsertServer(server); err != nil {
					p.errCh <- err
				}
			}
		case updateActionRemove:
			logrus.Debugf("proxy: removing server %s", u.backend.host)
			for _, server := range u.backend.servers {
				if err := u.backend.lb.RemoveServer(server); err != nil {
					p.errCh <- err
				}
			}
			if err := p.mux.Remove(host); err != nil {
				p.errCh <- err
			}
		}
	}
}

func (p *Proxy) Reload(apps []*applicationapi.App) error {
	// TODO: get current app list and trigger update
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
				server := &url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s:%d", svc.Name, ep.Port),
					Path:   "/",
				}
				appServers[ep.Host] = append(appServers[ep.Host], server)
			}
		}
	}

	updates := []*proxyUpdate{}

	// trigger updates
	for host, servers := range appServers {
		serverID, err := p.generateServerID(host)
		if err != nil {
			logrus.Errorf("proxy: error generating server id: %s", err)
			continue
		}

		b, exists := p.currentServers[serverID]
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

		update := &proxyUpdate{
			backend: b,
		}
		update.action = p.getUpdateAction(serverID)
		updates = append(updates, update)
		next[serverID] = b
	}

	// nothing changed; skip
	if reflect.DeepEqual(next, p.currentServers) {
		return nil
	}

	// notify updates
	for _, update := range updates {
		p.updateCh <- update
	}

	// prune
	p.pruneServers(next)
	// update current servers
	p.currentServers = next

	return nil
}

func (p *Proxy) Run() error {
	go p.updater()

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.config.HTTPPort),
		Handler: p.mux,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			logrus.Error(errors.Wrap(err, "proxy"))
		}
	}()

	return nil
}

func (p *Proxy) getUpdateAction(id string) updateAction {
	if _, ok := p.currentServers[id]; ok {
		return updateActionUpdate
	}

	return updateActionAdd
}

func (p *Proxy) generateServerID(v interface{}, extra ...interface{}) (string, error) {
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

func (p *Proxy) pruneServers(next map[string]*backend) {
	for k, b := range p.currentServers {
		if _, exists := next[k]; !exists {
			logrus.Debugf("proxy: %s not found; removing", k)
			p.updateCh <- &proxyUpdate{
				backend: b,
				action:  updateActionRemove,
			}
		}
	}
}

package memory

import api "github.com/stellarproject/radiant/api/v1"

type Memory struct {
	servers map[string]*api.Server
}

func NewMemory() *Memory {
	return &Memory{
		servers: make(map[string]*api.Server),
	}
}

func (m *Memory) Name() string {
	return "memory"
}

func (m *Memory) Add(host string, srv *api.Server) error {
	m.servers[host] = srv
	return nil
}

func (m *Memory) Remove(host string) error {
	if _, ok := m.servers[host]; ok {
		delete(m.servers, host)
	}
	return nil
}

func (m *Memory) Servers() ([]*api.Server, error) {
	servers := []*api.Server{}
	for _, srv := range m.servers {
		servers = append(servers, srv)
	}
	return servers, nil
}

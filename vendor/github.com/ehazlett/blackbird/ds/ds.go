package ds

import api "github.com/ehazlett/blackbird/api/v1"

type Datastore interface {
	// Name returns the name of the datastore
	Name() string
	// Add adds a server to the datastore
	Add(host string, srv *api.Server) error
	// Remove removes a server from the datastore
	Remove(host string) error
	// Servers returns all servers from the datastore
	Servers() ([]*api.Server, error)
}

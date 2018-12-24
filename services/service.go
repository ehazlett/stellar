package services

import (
	"google.golang.org/grpc"
)

type Type string

const (
	ApplicationService Type = "stellar.services.application.v1"
	ClusterService     Type = "stellar.services.cluster.v1"
	DatastoreService   Type = "stellar.services.datastore.v1"
	EventsService      Type = "stellar.services.events.v1"
	GatewayService     Type = "stellar.services.gateway.v1"
	HealthService      Type = "stellar.services.health.v1"
	NameserverService  Type = "stellar.services.nameserver.v1"
	NetworkService     Type = "stellar.services.network.v1"
	NodeService        Type = "stellar.services.node.v1"
	ProxyService       Type = "stellar.services.proxy.v1"
	SchedulerService   Type = "stellar.services.scheduler.v1"
	VersionService     Type = "stellar.services.version.v1"
)

// Service is the interface that all stellar services must implement
type Service interface {
	// ID returns the id of the service
	ID() string
	// Type returns the type that the service provides
	Type() Type
	// Register registers the service with the GRPC server
	Register(*grpc.Server) error
	// Requires returns a list of other service types needed by the service
	Requires() []Type
	// Start provides a mechanism to start service specific actions
	Start() error
	// stop provides a mechanism to stop the service
	Stop() error
}

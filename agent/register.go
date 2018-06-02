package agent

import (
	"fmt"

	"github.com/ehazlett/element/services"
)

// Register registers a GRPC service with the agent
func (a *Agent) Register(svc services.Service) error {
	id := svc.ID()
	if _, exists := a.registeredServices[id]; exists {
		return fmt.Errorf("service %s already registered", id)
	}
	svc.Register(a.grpcServer)
	return nil
}

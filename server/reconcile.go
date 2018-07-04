package server

import (
	"github.com/ehazlett/stellar/client"
	"github.com/sirupsen/logrus"
)

func (s *Server) reconcile() {
	c, err := client.NewClient(localNode.Addr)
	if err != nil {
		logrus.Error(err)
	}
	defer c.Close()

	// setup cluster routes
	if err := s.setupRoutes(); err != nil {
		logrus.Error(err)
	}
}

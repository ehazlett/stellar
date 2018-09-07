package application

import (
	"context"

	api "github.com/ehazlett/stellar/api/services/application/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) Create(ctx context.Context, req *api.CreateRequest) (*ptypes.Empty, error) {
	logrus.Debugf("creating application %s", req.Name)
	peers, err := s.agent.Peers()
	if err != nil {
		return empty, err
	}

	peerIdx := 0
	for _, service := range req.Services {
		// get random peer for deploy
		peer := peers[peerIdx]
		c, err := s.nodeClient(peer.Name)
		if err != nil {
			return empty, err
		}

		if err := c.Node().CreateContainer(req.Name, service); err != nil {
			return empty, err
		}

		// update proxy
		if err := c.Proxy().Reload(); err != nil {
			return empty, err
		}

		c.Close()

		// update peer index for next deploy
		if peerIdx > len(peers) {
			peerIdx = 0
		} else {
			peerIdx++
		}
	}
	return empty, nil
}

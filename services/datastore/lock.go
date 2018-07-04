package datastore

import (
	"context"
	"time"

	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) AcquireLock(ctx context.Context, _ *api.AcquireLockRequest) (*ptypes.Empty, error) {
	s.lock.Lock()

	go func() {
		select {
		case <-time.After(lockTimeout):
			logrus.Warnf("lock timeout occurred (%s)", lockTimeout)
			s.lock.Unlock()
		case <-s.lockChan:
			s.lock.Unlock()
		}
	}()

	return &ptypes.Empty{}, nil
}

func (s *service) ReleaseLock(ctx context.Context, _ *api.ReleaseLockRequest) (*ptypes.Empty, error) {
	s.lockChan <- true
	return &ptypes.Empty{}, nil
}

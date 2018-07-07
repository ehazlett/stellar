package datastore

import (
	"context"
	"time"

	api "github.com/ehazlett/stellar/api/services/datastore/v1"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

func (s *service) AcquireLock(ctx context.Context, req *api.AcquireLockRequest) (*ptypes.Empty, error) {
	s.lock.Lock()

	timeout, err := ptypes.DurationFromProto(req.Timeout)
	if err != nil {
		return empty, err
	}

	go func() {
		select {
		case <-time.After(timeout):
			logrus.Warnf("lock timeout occurred (%s)", timeout)
			s.lock.Unlock()
		case <-s.lockChan:
			s.lock.Unlock()
		}
	}()

	return empty, nil
}

func (s *service) ReleaseLock(ctx context.Context, _ *api.ReleaseLockRequest) (*ptypes.Empty, error) {
	s.lockChan <- true
	return empty, nil
}

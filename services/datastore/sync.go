package datastore

import (
	"context"
	"fmt"

	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/client"
)

type syncAction string

const (
	syncSet    syncAction = "set"
	syncDelete syncAction = "delete"
)

type syncOp struct {
	action syncAction
	bucket string
	key    string
	value  []byte
}

// syncWithPeers synchronizes the key among the other peers
func (s *service) syncWithPeers(ctx context.Context, op *syncOp) error {
	switch op.action {
	case syncSet:
		return s.peerSet(ctx, op.bucket, op.key, op.value)
	case syncDelete:
		return s.peerDelete(ctx, op.bucket, op.key)
	default:
		return fmt.Errorf("unsupported sync action")
	}

	return nil
}

func (s *service) peerSet(ctx context.Context, bucket, key string, value []byte) error {
	peers, err := s.agent.Peers()
	if err != nil {
		return err
	}
	for _, peer := range peers {
		ac, err := client.NewClient(peer.Addr)
		if err != nil {
			return err
		}

		if _, err := ac.DatastoreService().Set(ctx, &datastoreapi.SetRequest{
			Bucket: bucket,
			Key:    key,
			Value:  value,
			Sync:   false,
		}); err != nil {
			return err
		}

		ac.Close()
	}

	return nil
}

func (s *service) peerDelete(ctx context.Context, bucket, key string) error {
	peers, err := s.agent.Peers()
	if err != nil {
		return err
	}
	for _, peer := range peers {
		ac, err := client.NewClient(peer.Addr)
		if err != nil {
			return err
		}

		if _, err := ac.DatastoreService().Delete(ctx, &datastoreapi.DeleteRequest{
			Bucket: bucket,
			Key:    key,
			Sync:   false,
		}); err != nil {
			return err
		}

		ac.Close()
	}

	return nil
}

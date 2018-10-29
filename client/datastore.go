package client

import (
	"context"

	datastoreapi "github.com/ehazlett/stellar/api/services/datastore/v1"
	"github.com/ehazlett/stellar/api/types"
)

type datastore struct {
	client datastoreapi.DatastoreClient
}

func (d *datastore) Client() datastoreapi.DatastoreClient {
	return d.client
}

func (d *datastore) ID() (string, error) {
	ctx := context.Background()
	resp, err := d.client.Info(ctx, &datastoreapi.InfoRequest{})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *datastore) Get(bucket, key string) ([]byte, error) {
	ctx := context.Background()
	resp, err := d.client.Get(ctx, &datastoreapi.GetRequest{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		return nil, err
	}

	return resp.Data.Value, nil
}

func (d *datastore) CreateBucket(bucket string) error {
	ctx := context.Background()
	if _, err := d.client.CreateBucket(ctx, &datastoreapi.CreateBucketRequest{
		Bucket: bucket,
	}); err != nil {
		return err
	}
	return nil
}

func (d *datastore) Search(bucket, prefix string) ([]types.KeyValue, error) {
	ctx := context.Background()
	resp, err := d.client.Search(ctx, &datastoreapi.SearchRequest{
		Bucket: bucket,
		Prefix: prefix,
	})
	if err != nil {
		return nil, err
	}

	var data []types.KeyValue
	for _, kv := range resp.Data {
		data = append(data, types.KeyValue{
			Bucket: resp.Bucket,
			Key:    kv.Key,
			Value:  kv.Value,
		})
	}

	return data, nil
}

func (d *datastore) Set(bucket, key string, value []byte, sync bool) error {
	ctx := context.Background()
	if _, err := d.client.Set(ctx, &datastoreapi.SetRequest{
		Bucket: bucket,
		Key:    key,
		Value:  value,
		Sync:   sync,
	}); err != nil {
		return err
	}

	return nil
}

func (d *datastore) Delete(bucket, key string, sync bool) error {
	ctx := context.Background()
	if _, err := d.client.Delete(ctx, &datastoreapi.DeleteRequest{
		Bucket: bucket,
		Key:    key,
		Sync:   sync,
	}); err != nil {
		return err
	}

	return nil
}

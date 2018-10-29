package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/typeurl"
	nameserverapi "github.com/ehazlett/stellar/api/services/nameserver/v1"
	ptypes "github.com/gogo/protobuf/types"
)

type nameserver struct {
	client nameserverapi.NameserverClient
}

func (n *nameserver) ID() (string, error) {
	ctx := context.Background()
	resp, err := n.client.Info(ctx, &nameserverapi.InfoRequest{})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (n *nameserver) Create(rtype, name, value string, options interface{}) error {
	ctx := context.Background()
	t, err := recordType(rtype)
	if err != nil {
		return err
	}
	var opts *ptypes.Any
	if options != nil {
		o, err := typeurl.MarshalAny(options)
		if err != nil {
			return err
		}
		opts = o
	}
	if _, err := n.client.Create(ctx, &nameserverapi.CreateRequest{
		Name: name,
		Records: []*nameserverapi.Record{
			{
				Type:    t,
				Name:    name,
				Value:   value,
				Options: opts,
			},
		},
	}); err != nil {
		return err
	}
	return nil
}

func (n *nameserver) Lookup(query string) ([]*nameserverapi.Record, error) {
	ctx := context.Background()
	resp, err := n.client.Lookup(ctx, &nameserverapi.LookupRequest{
		Query: query,
	})
	if err != nil {
		return nil, err
	}
	return resp.Records, nil
}

func (n *nameserver) CreateRecords(name string, records []*nameserverapi.Record) error {
	ctx := context.Background()
	if _, err := n.client.Create(ctx, &nameserverapi.CreateRequest{
		Name:    name,
		Records: records,
	}); err != nil {
		return err
	}
	return nil
}

func (n *nameserver) List() ([]*nameserverapi.Record, error) {
	ctx := context.Background()
	resp, err := n.client.List(ctx, &nameserverapi.ListRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Records, nil
}

func (n *nameserver) Delete(rtype, name string) error {
	t, err := recordType(rtype)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if _, err := n.client.Delete(ctx, &nameserverapi.DeleteRequest{
		Type: t,
		Name: name,
	}); err != nil {
		return err
	}
	return nil
}

func recordType(rtype string) (nameserverapi.RecordType, error) {
	switch strings.ToUpper(rtype) {
	case "A":
		return nameserverapi.RecordType_A, nil
	case "CNAME":
		return nameserverapi.RecordType_CNAME, nil
	case "SRV":
		return nameserverapi.RecordType_SRV, nil
	case "TXT":
		return nameserverapi.RecordType_TXT, nil
	case "MX":
		return nameserverapi.RecordType_MX, nil
	default:
		return nameserverapi.RecordType_UNKNOWN, fmt.Errorf("unsupported record type %q", rtype)
	}
}

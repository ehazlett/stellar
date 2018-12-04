package server

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/stellarproject/radiant/ds"
	"github.com/stellarproject/radiant/ds/memory"
)

var (
	ErrUnsupportedDatastore = errors.New("unsupported datastore")
)

func getDatastore(uri string) (ds.Datastore, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "memory":
		return memory.NewMemory(), nil
	default:
		return nil, fmt.Errorf("u")
	}
}

package radiant

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	api "github.com/stellarproject/radiant/api/v1"
)

type AddOpts func(ctx context.Context, srv *api.Server) error

func WithPath(path string) AddOpts {
	return func(ctx context.Context, srv *api.Server) error {
		if path == "" {
			return fmt.Errorf("path cannot be empty")
		}
		srv.Path = path
		return nil
	}
}

func WithTLS(ctx context.Context, srv *api.Server) error {
	srv.TLS = true
	return nil
}

func WithPolicy(p api.Policy) AddOpts {
	return func(ctx context.Context, srv *api.Server) error {
		srv.Policy = p
		return nil
	}
}

func WithUpstreams(upstreams ...string) AddOpts {
	return func(ctx context.Context, srv *api.Server) error {
		srv.Upstreams = upstreams
		return nil
	}
}

func WithTimeouts(d time.Duration) AddOpts {
	return func(ctx context.Context, srv *api.Server) error {
		srv.Timeouts = types.DurationProto(d)
		return nil
	}
}

func WithPreset(preset string) AddOpts {
	return func(ctx context.Context, srv *api.Server) error {
		srv.Preset = preset
		return nil
	}
}

func WithServer(s *api.Server) AddOpts {
	return func(ctx context.Context, srv *api.Server) error {
		srv = s
		return nil
	}
}

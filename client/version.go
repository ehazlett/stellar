package client

import (
	"context"

	versionapi "github.com/ehazlett/stellar/api/services/version/v1"
	ptypes "github.com/gogo/protobuf/types"
)

type version struct {
	client versionapi.VersionClient
}

func (v *version) Version() (*versionapi.VersionResponse, error) {
	return v.client.Version(context.Background(), &ptypes.Empty{})
}

package types

import (
	"fmt"

	"github.com/containerd/typeurl"
)

const (
	nsOptionUrl = "stellar.io/services/nameserver/types"
)

func init() {
	typeurl.Register(&SRVOptions{}, nsOptionUrl+".SRVOptions")
}

type SRVOptions struct {
	Service  string
	Protocol string
	Priority uint16
	Weight   uint16
	Port     uint16
}

func (o *SRVOptions) String() string {
	return fmt.Sprintf("service=%s proto=%s priority=%d weight=%d port=%d",
		o.Service,
		o.Protocol,
		o.Priority,
		o.Weight,
		o.Port,
	)
}

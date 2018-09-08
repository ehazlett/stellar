package health

import (
	"time"

	ptypes "github.com/gogo/protobuf/types"
)

// Started returns the timestamp as time.Time
func (m *NodeHealth) Started() (time.Time, error) {
	return ptypes.TimestampFromProto(m.StartedAt)
}

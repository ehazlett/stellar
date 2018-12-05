package node

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/containerd/typeurl"
	"github.com/gogo/protobuf/types"
	radiantapi "github.com/stellarproject/radiant/api/v1"
)

func (m *Endpoint) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	for k, v := range v {
		switch k {
		case "service":
			m.Service = v.(string)
		case "host":
			m.Host = v.(string)
		case "port":
			m.Port = uint32(v.(float64))
		case "protocol":
			p, err := parseProtocol(v)
			if err != nil {
				return err
			}

			m.Protocol = p
		case "config":
			x := v.(map[string]interface{})
			switch m.Protocol {
			case Protocol_HTTP:
				// build any
				d := &radiantapi.Server{
					TLS: x["tls"].(bool),
				}
				if t, ok := x["timeouts"]; ok {
					tt, err := time.ParseDuration(t.(string))
					if err != nil {
						return err
					}
					d.Timeouts = types.DurationProto(tt)
				}
				if t, ok := x["proxy_try_duration"]; ok {
					tt, err := time.ParseDuration(t.(string))
					if err != nil {
						return err
					}
					d.ProxyTryDuration = types.DurationProto(tt)
				}
				if t, ok := x["proxy_fail_timeout"]; ok {
					tt, err := time.ParseDuration(t.(string))
					if err != nil {
						return err
					}
					d.ProxyFailTimeout = types.DurationProto(tt)
				}
				if t, ok := x["limits"]; ok {
					d.Limits = t.(string)
				}
				if t, ok := x["proxy_upstream_headers"]; ok {
					hdrs := map[string]string{}
					for kk, vv := range t.(map[string]interface{}) {
						hdrs[kk] = vv.(string)
					}
					d.ProxyUpstreamHeaders = hdrs
				}
				any, err := typeurl.MarshalAny(d)
				if err != nil {
					return err
				}
				m.EndpointConfig = any
			}
		}
	}

	return nil
}

func parseProtocol(v interface{}) (Protocol, error) {
	if v, ok := v.(string); ok {
		switch strings.ToLower(v) {
		case "tcp":
			return Protocol_TCP, nil
		case "udp":
			return Protocol_UDP, nil
		case "http":
			return Protocol_HTTP, nil
		default:
			return Protocol_UNKNOWN, fmt.Errorf("unknown protocol %s", v)
		}
	}
	if v, ok := v.(float64); ok {
		switch v {
		case float64(0):
			return Protocol_TCP, nil
		case float64(1):
			return Protocol_UDP, nil
		case float64(2):
			return Protocol_HTTP, nil
		default:
			return Protocol_UNKNOWN, fmt.Errorf("unknown protocol %v", v)
		}
	}

	return Protocol_UNKNOWN, fmt.Errorf("unknown protocol %v", v)
}

package node

import (
	"encoding/json"
	"fmt"
	"strings"
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
		case "tls":
			m.TLS = v.(bool)
		case "protocol":
			p, err := parseProtocol(v)
			if err != nil {
				return err
			}

			m.Protocol = p
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

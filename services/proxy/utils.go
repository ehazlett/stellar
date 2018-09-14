package proxy

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

func checkConnection(upstream string, timeout time.Duration) (time.Duration, error) {
	zero := time.Millisecond * 0
	endpoint, err := url.Parse(upstream)
	if err != nil {
		return zero, err
	}

	switch endpoint.Scheme {
	case "tcp":
		c, err := net.DialTimeout("tcp", endpoint.Host, timeout)
		if err != nil {
			return zero, err
		}
		defer c.Close()
		start := time.Now()
		if _, err := c.Read([]byte{}); err != nil {
			return zero, err
		}

		latency := time.Now().Sub(start)
		return latency, nil
	case "http", "https":
		c := &http.Client{
			Timeout: timeout,
		}
		start := time.Now()
		if _, err := c.Get(endpoint.Scheme + "://" + endpoint.Host); err != nil {
			return zero, err
		}
		latency := time.Now().Sub(start)
		return latency, nil
	}

	return zero, nil
}

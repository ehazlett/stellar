// +build !windows,!linux

package health

import "errors"

var ErrPlatformUnsupported = errors.New("platform is unsupported")

func OSInfo() (*Info, error) {
	return nil, ErrPlatformUnsupported
}

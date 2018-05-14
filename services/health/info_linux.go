// +build linux

package health

import (
	"bytes"

	"golang.org/x/sys/unix"
)

func OSInfo() (*Info, error) {
	uts := &unix.Utsname{}
	if err := unix.Uname(uts); err != nil {
		return nil, err
	}

	return &Info{
		OSName:    string(uts.Sysname[:bytes.IndexByte(uts.Sysname[:], 0)]),
		OSVersion: string(uts.Release[:bytes.IndexByte(uts.Release[:], 0)]),
	}, nil
}

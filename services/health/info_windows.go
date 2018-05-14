// +build windows

package health

import "golang.org/x/sys/windows/registry"

func OSInfo() (*Info, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	name, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return nil, err
	}

	build, _, err := k.GetStringValue("CurrentBuild")
	if err != nil {
		return nil, err
	}
	return &Info{
		OSName:    name,
		OSVersion: build,
	}, nil
}

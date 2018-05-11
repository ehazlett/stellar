package version

import (
	"runtime"
	"testing"
)

func TestFullVersion(t *testing.T) {
	version := FullVersion()

	expected := Name + "/" + Version + Build + " (" + GitCommit + ") " + runtime.GOOS + "/" + runtime.GOARCH

	if version != expected {
		t.Fatalf("invalid version returned: %s", version)
	}
}

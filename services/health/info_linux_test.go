package health

import (
	"testing"
)

func TestOSInfo(t *testing.T) {
	v, err := OSInfo()
	if err != nil {
		t.Fatal(err)
	}

	if v.OSName != "Linux" {
		t.Fatalf("expected %q; received %q", "Linux", v.OSName)
	}

	if v.OSVersion == "" {
		t.Fatal("expected os version")
	}
}

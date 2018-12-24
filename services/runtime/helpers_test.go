package runtime

import (
	"fmt"
	"testing"
)

func TestGetNameDuplicates(t *testing.T) {
	names := map[string]struct{}{}
	for i := 0; i < 1024; i++ {
		name := getName(fmt.Sprintf("test-%d", i))
		if _, exists := names[name]; exists {
			t.Errorf("generated duplicate name %s", name)
		}
	}
}

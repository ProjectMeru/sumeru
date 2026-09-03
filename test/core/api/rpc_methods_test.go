package api_test

import (
	"testing"
	"sumeru/core/server/api"
)


func TestPublicMethods_includesReadGroupAndCall(t *testing.T) {
	for _, m := range []string{"read_group", "call"} {
		if !api.PublicMethods[m] {
			t.Fatalf("expected %q in api.PublicMethods", m)
		}
	}
}

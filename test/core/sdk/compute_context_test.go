package sdk_test

import (
	"testing"

	"sumeru/core/sdk"
)

func TestNewComputeContext(t *testing.T) {
	c := sdk.NewComputeContext(42)
	if c.ID != 42 {
		t.Fatalf("ID=%d", c.ID)
	}
}

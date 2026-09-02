package digest_test

import (
	"context"
	"testing"

	"sumeru/addons/digest"
)

func TestComputeKPIUnknownCode(t *testing.T) {
	_, err := digest.ComputeKPI(context.Background(), "missing.code")
	if err == nil {
		t.Fatal("expected error for unknown compute code")
	}
}

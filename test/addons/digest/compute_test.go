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

func TestComputeKPIMissingModelReturnsZero(t *testing.T) {
	v, err := digest.ComputeKPI(context.Background(), "crm.lead_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 0 {
		t.Fatalf("expected 0 when crm.lead absent, got %v", v)
	}
}

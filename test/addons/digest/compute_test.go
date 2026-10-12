package digest_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/addons/digest"
)

func TestComputeKPI_unknownCode(t *testing.T) {
	_, err := digest.ComputeKPI(context.Background(), "not.registered")
	if err == nil || !strings.Contains(err.Error(), "unknown compute code") {
		t.Fatalf("expected unknown code error, got %v", err)
	}
}

func TestComputeKPI_missingModelReturnsZero(t *testing.T) {
	val, err := digest.ComputeKPI(context.Background(), "crm.lead_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Fatalf("got %v want 0 without crm module", val)
	}
}

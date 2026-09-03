package digest_test

import (
	"context"
	"testing"

	"sumeru/addons/digest"
)
func TestComputeKPI_table(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		code    string
		wantErr bool
	}{
		{"missing.code", true},
		{"crm.lead_count", false},
		{"crm.lead_won_count", false},
		{"crm.opportunity_count", false},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			v, err := digest.ComputeKPI(ctx, tc.code)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if v != 0 {
				t.Fatalf("without registry expected 0, got %v", v)
			}
		})
	}
}

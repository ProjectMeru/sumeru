package modelmeta_test

import (
	"testing"

	"sumeru/core/modelmeta"
)

func TestExtractDefaultFromSelectionTail_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in          string
		wantSel     string
		wantDefault string
	}{
		{"a:A,b:B,default=a", "a:A,b:B", "a"},
		{"x:X", "x:X", ""},
		{"", "", ""},
		{"only:Only,default=only", "only:Only", "only"},
	}
	for _, tc := range tests {
		sel, def := modelmeta.ExtractDefaultFromSelectionTailForTest(tc.in)
		if def != tc.wantDefault {
			t.Fatalf("ExtractDefaultFromSelectionTailForTest(%q) default=%q want %q", tc.in, def, tc.wantDefault)
		}
		if tc.wantSel != "" && sel == "" {
			t.Fatalf("ExtractDefaultFromSelectionTailForTest(%q) trimmed selection empty", tc.in)
		}
	}
}

func TestPeelSelectionTag_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tag        string
		wantBody   string
		wantSelect string
	}{
		{"required,selection=a:A", "required", "a:A"},
		{"selection=x:X,y:Y", "", "x:X,y:Y"},
		{"readonly", "readonly", ""},
	}
	for _, tc := range tests {
		body, sel := modelmeta.PeelSelectionTagForTest(tc.tag)
		if body != tc.wantBody || sel != tc.wantSelect {
			t.Fatalf("PeelSelectionTagForTest(%q) = (%q, %q) want (%q, %q)", tc.tag, body, sel, tc.wantBody, tc.wantSelect)
		}
	}
}

func TestFieldNameFromGo_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		goName string
		want   string
	}{
		{"CompanyID", "company_id"},
		{"Name", "name"},
		{"LineIDs", "line_ids"},
	}
	for _, tc := range tests {
		if got := modelmeta.FieldNameFromGo(tc.goName); got != tc.want {
			t.Errorf("FieldNameFromGo(%q) = %q want %q", tc.goName, got, tc.want)
		}
	}
}

func TestHeuristicGoName_table(t *testing.T) {
	t.Parallel()
	if got := modelmeta.HeuristicGoName("sale.order"); got != "SaleOrder" {
		t.Fatalf("HeuristicGoName: %q", got)
	}
	if got := modelmeta.HeuristicGoName(""); got != "" {
		t.Fatalf("empty model: %q", got)
	}
}

package report_test

import (
	"strings"
	"testing"

	"sumeru/core/report"
)

func TestExportTemplatePDFMinimal(t *testing.T) {
	data, err := report.ExportTemplatePDF(report.TemplatePDFInput{
		Title:    "Sales Order SO-001",
		Subtitle: "Acme Corp",
		Sections: []report.TemplatePDFSection{
			{Heading: "Summary", Body: "Total amount: 1,200.00"},
		},
		TableHead: []string{"Product", "Qty"},
		TableRows: [][]string{{"Widget", "3"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 || !strings.HasPrefix(string(data), "%PDF") {
		t.Fatalf("expected PDF bytes, got len=%d", len(data))
	}
}

func TestRenderTemplatePDFText(t *testing.T) {
	text, err := report.RenderTemplatePDFText(report.TemplatePDFInput{
		Title: "Invoice",
		Sections: []report.TemplatePDFSection{
			{Heading: "Due", Body: "Net 30"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Invoice") || !strings.Contains(text, "Net 30") {
		t.Fatalf("unexpected preview: %q", text)
	}
}

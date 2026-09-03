package report_test

import (
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/report"
)

func TestCapabilitiesFromView_noReport(t *testing.T) {
	caps := report.CapabilitiesFromView(&parser.View{Type: "list"})
	if caps.HasDownload() || caps.BulkUpload {
		t.Fatalf("expected disabled caps, got %+v", caps)
	}
}

func TestBulkTemplateCSV(t *testing.T) {
	t.Skip("requires registered model in ORM")
}

func TestImportFlashMessage(t *testing.T) {
	got := report.ImportFlashMessage(report.ImportResult{Created: 5, Updated: 2, Skipped: 1})
	want := "imported_5_updated_2_skipped_1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestParseMappingJSON(t *testing.T) {
	m, err := report.ParseMappingJSON(`{"Name":"name","Email":"email"}`)
	if err != nil {
		t.Fatal(err)
	}
	if m["Name"] != "name" {
		t.Fatalf("mapping = %#v", m)
	}
}

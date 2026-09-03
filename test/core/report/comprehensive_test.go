package report_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/report"
)

func TestCapabilitiesFromView_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		view *parser.View
		check func(t *testing.T, caps report.Capabilities)
	}{
		{
			name: "nil view",
			view: nil,
			check: func(t *testing.T, caps report.Capabilities) {
				if caps.HasDownload() || caps.BulkUpload {
					t.Fatal("nil view caps should be empty")
				}
			},
		},
		{
			name: "report element",
			view: func() *parser.View {
				v, _ := parser.ParseViewFromArch(`<list><report download="csv,xlsx" upload="bulk" pdf_sizes="a4" modes="create"/></list>`)
				return v
			}(),
			check: func(t *testing.T, caps report.Capabilities) {
				if !caps.HasDownload() || !caps.BulkUpload {
					t.Fatalf("report element: %+v", caps)
				}
				if len(caps.PDFSizes) == 0 || len(caps.BulkModes) == 0 {
					t.Fatalf("defaults: %+v", caps)
				}
			},
		},
		{
			name: "header widgets",
			view: func() *parser.View {
				v, _ := parser.ParseViewFromArch(`<view type="list"><header><widget type="report_download" formats="pdf"/><widget type="bulk_upload" modes="upsert"/></header><field name="x"/></view>`)
				return v
			}(),
			check: func(t *testing.T, caps report.Capabilities) {
				if !caps.HasDownload() || !caps.BulkUpload {
					t.Fatalf("header widgets: %+v", caps)
				}
			},
		},
		{
			name: "bulk upload attr",
			view: func() *parser.View {
				v, _ := parser.ParseViewFromArch(`<view type="list" bulk_upload="1"><field name="x"/></view>`)
				return v
			}(),
			check: func(t *testing.T, caps report.Capabilities) {
				if !caps.BulkUpload {
					t.Fatal("bulk_upload attr")
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caps := report.CapabilitiesFromView(tc.view)
			if tc.check != nil {
				tc.check(t, caps)
			}
		})
	}
}

func TestDefaultDocumentLayout(t *testing.T) {
	t.Parallel()
	layout := report.DefaultDocumentLayout("")
	if layout.CompanyName != "Sumeru" || layout.PaperFormat != report.PageSizeA4 {
		t.Fatalf("default layout: %+v", layout)
	}
	layout = report.DefaultDocumentLayout("  Acme Corp  ")
	if layout.CompanyName != "Acme Corp" {
		t.Fatalf("company name: %+v", layout)
	}
}

func TestParseFieldsAndActionID(t *testing.T) {
	t.Parallel()
	if got := report.ParseFieldsParam(" id , name , ,email "); len(got) != 3 {
		t.Fatalf("ParseFieldsParam: %v", got)
	}
	if report.ParseFieldsParam("  ") != nil {
		t.Fatal("ParseFieldsParam blank")
	}
	if got := report.ParseActionIDParam(" 42 "); got != 42 {
		t.Fatalf("ParseActionIDParam: %d", got)
	}
}

func TestMappingFormURL(t *testing.T) {
	t.Parallel()
	withAction := report.MappingFormURL(7, 10)
	if !strings.Contains(withAction, "action=10") || !strings.Contains(withAction, "id=7") {
		t.Fatalf("with action: %q", withAction)
	}
	noAction := report.MappingFormURL(7, 0)
	if !strings.Contains(noAction, "model="+report.BulkModelName) {
		t.Fatalf("no action: %q", noAction)
	}
}

func TestImportFlashMessage_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   report.ImportResult
		want string
	}{
		{in: report.ImportResult{Created: 1}, want: "imported_1_updated_0_skipped_0"},
		{in: report.ImportResult{Created: 2, Updated: 3, Skipped: 1}, want: "imported_2_updated_3_skipped_1"},
		{in: report.ImportResult{Updated: 5}, want: "imported_0_updated_5_skipped_0"},
	}
	for _, tc := range tests {
		if got := report.ImportFlashMessage(tc.in); got != tc.want {
			t.Fatalf("ImportFlashMessage(%+v) = %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseMappingJSON_errors(t *testing.T) {
	t.Parallel()
	if _, err := report.ParseMappingJSON(""); err != nil {
		t.Fatalf("empty mapping: %v", err)
	}
	if _, err := report.ParseMappingJSON("not-json"); err == nil {
		t.Fatal("invalid json should error")
	}
	m, err := report.ParseMappingJSON(`{"A":"a","B":"b"}`)
	if err != nil || m["A"] != "a" {
		t.Fatalf("mapping: %v err=%v", m, err)
	}
}

func TestRegisterCellFormatter(t *testing.T) {
	t.Parallel()
	report.RegisterCellFormatter("test.export.fmt", func(_ context.Context, _, field string, _ interface{}) string {
		if field == "name" {
			return "FORMATTED"
		}
		return ""
	})
	// Indirectly exercised via export paths in other tests; register empty is no-op
	report.RegisterCellFormatter("", nil)
	report.RegisterCellFormatter("x", nil)
}

func TestExportTemplatePDF_minimal(t *testing.T) {
	data, err := report.ExportTemplatePDF(report.TemplatePDFInput{
		Title:    "Invoice",
		Subtitle: "Draft",
		Sections: []report.TemplatePDFSection{{Heading: "Notes", Body: "Thanks"}},
	})
	if err != nil || len(data) == 0 {
		t.Fatalf("ExportTemplatePDF: len=%d err=%v", len(data), err)
	}
}

func TestExportTemplatePDFWithLayout(t *testing.T) {
	layout := report.DefaultDocumentLayout("Co")
	layout.PaperFormat = report.PageSizeLetter
	data, err := report.ExportTemplatePDFWithLayout(report.TemplatePDFInput{Title: "T"}, layout)
	if err != nil || len(data) == 0 {
		t.Fatalf("ExportTemplatePDFWithLayout: err=%v", err)
	}
}

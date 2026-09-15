package report_test

import (
	"strings"
	"testing"

	"sumeru/core/report"
)

func TestRenderHTMLTemplate(t *testing.T) {
	report.RegisterHTMLTemplate("test_report_html", `<h1>Title</h1><p>Body {{.Report.model}}</p>`)
	in, err := report.RenderHTMLTemplate("test_report_html", report.DocumentRenderContext{
		Report: map[string]interface{}{"model": "core.company"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if in.Title != "Title" {
		t.Fatalf("title %q", in.Title)
	}
	if len(in.Sections) == 0 || !strings.Contains(in.Sections[0].Body, "core.company") {
		t.Fatalf("sections %+v", in.Sections)
	}
}

func TestRenderHTMLTemplate_missing(t *testing.T) {
	_, err := report.RenderHTMLTemplate("missing_template_xyz", report.DocumentRenderContext{})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err %v", err)
	}
}

func TestExportTemplatePDFFromDocumentArch(t *testing.T) {
	arch := `<document><title>T</title><p>Line</p></document>`
	in, err := report.TemplatePDFInputFromDocumentArch(arch, report.DocumentRenderContext{}, report.PageSizeA4)
	if err != nil {
		t.Fatal(err)
	}
	pdf, err := report.ExportTemplatePDF(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(pdf) < 100 {
		t.Fatalf("pdf too short: %d", len(pdf))
	}
}

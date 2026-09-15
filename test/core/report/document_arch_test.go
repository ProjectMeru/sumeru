package report_test

import (
	"strings"
	"testing"

	"sumeru/core/report"
)

func TestTemplatePDFInputFromDocumentArch(t *testing.T) {
	arch := `<document><title>Hello</title><section heading="H"><field name="name"/></section></document>`
	ctx := report.DocumentRenderContext{
		Record: map[string]interface{}{"name": "Acme"},
	}
	in, err := report.TemplatePDFInputFromDocumentArch(arch, ctx, report.PageSizeA4)
	if err != nil {
		t.Fatal(err)
	}
	if in.Title != "Hello" {
		t.Fatalf("title: %q", in.Title)
	}
	if len(in.Sections) == 0 || !strings.Contains(in.Sections[0].Body, "Acme") {
		t.Fatalf("sections: %+v", in.Sections)
	}
}

func TestTemplatePDFInputFromDocumentArch_precedenceEmpty(t *testing.T) {
	_, err := report.TemplatePDFInputFromDocumentArch("", report.DocumentRenderContext{}, report.PageSizeA4)
	if err == nil {
		t.Fatal("expected error for empty arch")
	}
}

func TestBuildDocumentPDF_archOverTemplate(t *testing.T) {
	t.Skip("requires ORM")
}

package report_test

import (
	"strings"
	"testing"

	"sumeru/core/engine/viewinherit"
	"sumeru/core/report"
)

func TestDocumentArchInheritChain(t *testing.T) {
	base := `<document><title>Base</title><section heading="Details"><field name="name"/></section></document>`
	frag := `<xpath expr="//section[@heading='Details']" position="inside"><p>Inherited</p></xpath>`
	out, err := viewinherit.ApplyInheritArch(base, frag)
	if err != nil {
		t.Fatal(err)
	}
	in, err := report.TemplatePDFInputFromDocumentArch(out, report.DocumentRenderContext{
		Record: map[string]interface{}{"name": "X"},
	}, report.PageSizeA4)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(in.Sections[0].Body, "Inherited") {
		t.Fatalf("body %q", in.Sections[0].Body)
	}
}

func TestStoreReportPDFAttachment_requiresID(t *testing.T) {
	t.Skip("requires ORM")
}

package report_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/report"
)

func TestCapabilitiesXLSXFormat(t *testing.T) {
	arch := `<list><report download="csv,xlsx,pdf"/><field name="name"/></list>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	caps := report.CapabilitiesFromView(v)
	if len(caps.DownloadFormats) != 3 {
		t.Fatalf("formats = %v", caps.DownloadFormats)
	}
}

func TestMinimalXLSXZipStructure(t *testing.T) {
	data, err := report.ExportXLSXForTest([]string{"Name"}, [][]string{{"Acme"}})
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "sheet1.xml") {
			found = true
		}
	}
	if !found {
		t.Fatal("missing worksheet in xlsx")
	}
}

func TestXLSXColumnAA(t *testing.T) {
	headers := make([]string, 27)
	for i := range headers {
		headers[i] = fmt.Sprintf("Col%d", i)
	}
	sheet := report.SheetXMLForTest(headers, nil)
	if !strings.Contains(sheet, `r="AA1"`) {
		t.Fatalf("expected AA1 cell ref in %q", sheet)
	}
}

func TestXLSXEscapesSpecialChars(t *testing.T) {
	sheet := report.SheetXMLForTest([]string{"Label"}, [][]string{{`<script>&"'`}})
	if !strings.Contains(sheet, "&lt;script&gt;") || !strings.Contains(sheet, "&amp;") {
		t.Fatalf("unexpected sheet xml: %q", sheet)
	}
}

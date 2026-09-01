package report_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"sumeru/core/report"
)

func TestSanitizeSpreadsheetCell(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Acme", "Acme"},
		{"=1+1", "\t=1+1"},
		{"+cmd", "\t+cmd"},
		{"-10", "-10"},
		{"-42.5", "-42.5"},
		{"-1+1", "\t-1+1"},
		{"@SUM(A1)", "\t@SUM(A1)"},
	}
	for _, tc := range tests {
		got := report.SanitizeSpreadsheetCellForTest(tc.in)
		if got != tc.want {
			t.Fatalf("sanitize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestExportXLSXSanitizesFormulaPrefix(t *testing.T) {
	data, err := report.ExportXLSXForTest([]string{"Name"}, [][]string{{"=1+1"}})
	if err != nil {
		t.Fatal(err)
	}
	sheet, err := readZipFile(data, "xl/worksheets/sheet1.xml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sheet, "\t=1+1") {
		t.Fatalf("sheet xml = %q", sheet)
	}
}

func TestWriteCSVSanitizesFormulaPrefix(t *testing.T) {
	data, err := report.WriteCSVForTest([]string{"Name"}, [][]string{{"=1+1"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "\t=1+1") {
		t.Fatalf("csv = %q", string(data))
	}
}

func readZipFile(data []byte, name string) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		return string(body), nil
	}
	return "", io.EOF
}

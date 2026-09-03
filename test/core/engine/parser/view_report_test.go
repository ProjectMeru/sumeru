package parser_test

import (
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/report"
)

func TestReportElementOnListArch(t *testing.T) {
	arch := `<list><report download="csv,pdf" upload="bulk"/><field name="name"/></list>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	caps := report.CapabilitiesFromView(v)
	if !caps.HasDownload() {
		t.Fatal("expected download enabled")
	}
	if len(caps.DownloadFormats) != 2 {
		t.Fatalf("formats = %v", caps.DownloadFormats)
	}
	if !caps.BulkUpload {
		t.Fatal("expected bulk upload")
	}
}

func TestReportViewAttributes(t *testing.T) {
	arch := `<view type="list" model="product.product" report_download="csv" bulk_upload="1"><field name="name"/></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	caps := report.CapabilitiesFromView(v)
	if !caps.HasDownload() || caps.DownloadFormats[0] != "csv" {
		t.Fatalf("caps = %+v", caps)
	}
	if !caps.BulkUpload {
		t.Fatal("expected bulk upload from attr")
	}
}

func TestReportHeaderWidget(t *testing.T) {
	arch := `<view type="form" model="crm.lead"><header><widget type="report_download" formats="pdf"/><widget type="bulk_upload"/></header></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	caps := report.CapabilitiesFromView(v)
	if len(caps.DownloadFormats) != 1 || caps.DownloadFormats[0] != "pdf" {
		t.Fatalf("formats = %v", caps.DownloadFormats)
	}
	if !caps.BulkUpload {
		t.Fatal("expected bulk from widget")
	}
}

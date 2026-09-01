package report

import (
	"strings"

	"sumeru/core/engine/parser"
)

// CapabilitiesFromView merges <report>, view attrs, and header widgets.
func CapabilitiesFromView(v *parser.View) Capabilities {
	if v == nil {
		return Capabilities{}
	}
	var caps Capabilities
	if v.Report != nil {
		mergeCaps(&caps, v.Report.Download, v.Report.Upload, v.Report.PDFSizes, v.Report.Modes)
	}
	if v.ReportDownload != "" {
		mergeCaps(&caps, v.ReportDownload, "", v.ReportPDFSizes, v.ReportBulkModes)
	}
	if parser.IsTruthyAttr(v.BulkUpload) {
		caps.BulkUpload = true
	}
	if v.Header != nil {
		for _, w := range v.Header.Widget {
			switch strings.ToLower(strings.TrimSpace(w.Type)) {
			case "report_download":
				mergeCaps(&caps, w.Formats, "", w.PDFSizes, "")
			case "bulk_upload":
				caps.BulkUpload = true
				if w.Modes != "" {
					caps.BulkModes = splitCSVList(w.Modes)
				}
			}
		}
	}
	normalizeCaps(&caps)
	return caps
}

func mergeCaps(caps *Capabilities, download, upload, pdfSizes, modes string) {
	for _, f := range splitCSVList(download) {
		if f == "csv" || f == "pdf" || f == "xlsx" {
			caps.DownloadFormats = appendUnique(caps.DownloadFormats, f)
		}
	}
	if strings.EqualFold(strings.TrimSpace(upload), "bulk") || parser.IsTruthyAttr(upload) {
		caps.BulkUpload = true
	}
	if pdfSizes != "" {
		caps.PDFSizes = splitCSVList(pdfSizes)
	}
	if modes != "" {
		caps.BulkModes = splitCSVList(modes)
	}
}

func normalizeCaps(caps *Capabilities) {
	if caps.BulkUpload && len(caps.BulkModes) == 0 {
		caps.BulkModes = []string{ImportModeCreate, ImportModeUpsert}
	}
	if caps.HasDownload() && len(caps.PDFSizes) == 0 {
		caps.PDFSizes = []string{PageSizeA4, PageSizeLegal, PageSizeLetter}
	}
}

func splitCSVList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var out []string
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func appendUnique(list []string, items ...string) []string {
	seen := map[string]struct{}{}
	for _, x := range list {
		seen[x] = struct{}{}
	}
	for _, x := range items {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		list = append(list, x)
	}
	return list
}


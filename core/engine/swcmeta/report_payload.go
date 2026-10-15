package swcmeta

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
	"sumeru/core/report"
)

// Master-data models always expose Reports import/export in workspace chrome.
var masterDataReportModels = map[string]struct{}{
	"core.user":          {},
	"core.company":       {},
	"core.group":         {},
	"core.lang":          {},
	"core.country":       {},
	"core.country.state": {},
	"core.city":          {},
	"core.user.apikey":   {},
}

// ApplyReportToArch sets arch.Report from the active view, sibling action views, or master-data defaults.
func ApplyReportToArch(
	ctx context.Context,
	arch *ViewArch,
	model string,
	activeView *parser.View,
	actionViewModes []string,
) {
	if arch == nil {
		return
	}
	if activeView != nil && activeView.Report != nil && report.ReportElementInvisible(activeView.Report.Invisible) {
		arch.Report = nil
		return
	}
	if arch.Report != nil {
		return
	}
	if caps := mergeReportFromSiblingViews(ctx, model, actionViewModes); caps.HasDownload() || caps.BulkUpload {
		arch.Report = reportMetaFromCapabilities(caps)
		return
	}
	if _, ok := masterDataReportModels[strings.TrimSpace(model)]; ok {
		arch.Report = defaultMasterDataReportMeta()
	}
}

// ExportEntitlementsForModel merges report download/upload capabilities from action view modes,
// including master-data defaults used in workspace chrome.
func ExportEntitlementsForModel(ctx context.Context, model string, actionViewModes []string) report.Capabilities {
	caps := mergeReportFromSiblingViews(ctx, model, actionViewModes)
	if caps.HasDownload() || caps.BulkUpload {
		return caps
	}
	if _, ok := masterDataReportModels[strings.TrimSpace(model)]; ok {
		mergeDefaultMasterDataCaps(&caps)
		normalizeReportCapabilities(&caps)
	}
	return caps
}

func mergeReportFromSiblingViews(ctx context.Context, model string, modes []string) report.Capabilities {
	model = strings.TrimSpace(model)
	if model == "" {
		return report.Capabilities{}
	}
	var merged report.Capabilities
	seen := map[string]struct{}{}
	for _, mode := range modes {
		mode = strings.TrimSpace(strings.ToLower(mode))
		if mode == "" {
			continue
		}
		if _, ok := seen[mode]; ok {
			continue
		}
		seen[mode] = struct{}{}
		viewData, err := orm.FindUIDefaultView(ctx, model, mode)
		if err != nil || viewData == nil {
			continue
		}
		archText := strings.TrimSpace(orm.AsString(viewData["arch"]))
		if archText == "" {
			continue
		}
		parsed, err := parser.ParseViewFromArch(archText)
		if err != nil || parsed == nil {
			continue
		}
		merged = unionReportCapabilities(merged, report.CapabilitiesFromView(parsed))
	}
	return merged
}

func unionReportCapabilities(a, b report.Capabilities) report.Capabilities {
	out := a
	out.DownloadFormats = appendUniqueStrings(out.DownloadFormats, b.DownloadFormats...)
	if b.BulkUpload {
		out.BulkUpload = true
	}
	out.PDFSizes = appendUniqueStrings(out.PDFSizes, b.PDFSizes...)
	out.BulkModes = appendUniqueStrings(out.BulkModes, b.BulkModes...)
	normalizeReportCapabilities(&out)
	return out
}

func appendUniqueStrings(list []string, items ...string) []string {
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

func normalizeReportCapabilities(caps *report.Capabilities) {
	if caps.BulkUpload && 0 == len(caps.BulkModes) {
		caps.BulkModes = []string{report.ImportModeCreate, report.ImportModeUpsert}
	}
	if caps.HasDownload() && 0 == len(caps.PDFSizes) {
		caps.PDFSizes = []string{report.PageSizeA4, report.PageSizeLegal, report.PageSizeLetter}
	}
}

func reportMetaFromCapabilities(caps report.Capabilities) *ReportMeta {
	return &ReportMeta{
		Download:  caps.HasDownload(),
		Upload:    caps.BulkUpload,
		Formats:   strings.Join(caps.DownloadFormats, ","),
		PDFSizes:  strings.Join(caps.PDFSizes, ","),
		BulkModes: strings.Join(caps.BulkModes, ","),
	}
}

func defaultMasterDataReportMeta() *ReportMeta {
	caps := report.Capabilities{}
	mergeDefaultMasterDataCaps(&caps)
	normalizeReportCapabilities(&caps)
	return reportMetaFromCapabilities(caps)
}

func mergeDefaultMasterDataCaps(caps *report.Capabilities) {
	for _, fmt := range []string{"csv", "pdf"} {
		caps.DownloadFormats = appendUniqueStrings(caps.DownloadFormats, fmt)
	}
	caps.BulkUpload = true
}

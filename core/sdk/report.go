package sdk

import (
	"context"

	"sumeru/core/report"
)

// Report page size constants (re-exported from core/report).
const (
	PageSizeA4       = report.PageSizeA4
	PageSizeLegal    = report.PageSizeLegal
	PageSizeLetter   = report.PageSizeLetter
	ImportModeCreate = report.ImportModeCreate
	ImportModeUpsert = report.ImportModeUpsert
)

// ReportCellFormatter customizes export cell text for a model field.
type ReportCellFormatter = report.CellFormatter

// ExportCSVInput configures CSV export via the SDK.
type ExportCSVInput = report.ExportCSVInput

// ExportPDFInput configures PDF export via the SDK.
type ExportPDFInput = report.ExportPDFInput

// PreviewBulkImportInput configures bulk import preview.
type PreviewBulkImportInput = report.PreviewBulkImportInput

// ExecuteBulkImportInput configures bulk import execution.
type ExecuteBulkImportInput = report.ExecuteBulkImportInput

// ImportResult holds bulk import counts.
type ImportResult = report.ImportResult

// PreviewResult holds bulk preview validation.
type PreviewResult = report.PreviewResult

// Capabilities describes report features enabled on a parsed view.
type Capabilities = report.Capabilities

// ExportCSV delegates to core/report.
func ExportCSV(ctx context.Context, in ExportCSVInput) ([]byte, error) {
	return report.ExportCSV(ctx, in)
}

// ExportPDF delegates to core/report.
func ExportPDF(ctx context.Context, in ExportPDFInput) ([]byte, error) {
	return report.ExportPDF(ctx, in)
}

// BulkTemplateCSV delegates to core/report.
func BulkTemplateCSV(modelName string, fields []string) ([]byte, error) {
	return report.BulkTemplateCSV(modelName, fields)
}

// PreviewBulkImport delegates to core/report.
func PreviewBulkImport(ctx context.Context, in PreviewBulkImportInput) (PreviewResult, error) {
	return report.PreviewBulkImport(ctx, in)
}

// ExecuteBulkImport delegates to core/report.
func ExecuteBulkImport(ctx context.Context, in ExecuteBulkImportInput) (ImportResult, error) {
	return report.ExecuteBulkImport(ctx, in)
}

// RegisterReportCellFormatter registers an export cell formatter for a model.
func RegisterReportCellFormatter(model string, fn ReportCellFormatter) {
	report.RegisterCellFormatter(model, fn)
}

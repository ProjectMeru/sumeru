// Package report implements CSV/PDF export and bulk CSV import for workspace views.
package report

const (
	PageSizeA4     = "a4"
	PageSizeLegal  = "legal"
	PageSizeLetter = "letter"

	ImportModeCreate = "create"
	ImportModeUpsert = "upsert"

	BulkModelName = "sys.bulk.import"

	DirectionImport = "import"
	DirectionExport = "export"

	BulkStateDraft     = "draft"
	BulkStateQueued    = "queued"
	BulkStateRunning   = "running"
	BulkStateDone      = "done"
	BulkStateFailed    = "failed"
	BulkStateCancelled = "cancelled"

	BulkJobCronCode = "platform.bulk_job"

	maxExportRows           = 500
	maxSyncExportRows       = 500
	maxPreviewRows          = 50
	asyncImportRowThreshold = 500
)

// Capabilities describes report download and bulk upload enabled on a view.
type Capabilities struct {
	DownloadFormats []string
	BulkUpload      bool
	PDFSizes        []string
	BulkModes       []string
}

// HasDownload reports whether any download format is enabled.
func (c Capabilities) HasDownload() bool {
	return len(c.DownloadFormats) > 0
}

// ImportResult counts rows affected by bulk import.
type ImportResult struct {
	Created int
	Updated int
	Skipped int
}

// PreviewRow is one validated preview line.
type PreviewRow struct {
	RowNum int
	Values map[string]interface{}
	Errors []string
}

// PreviewResult is the outcome of mapping + validation before import.
type PreviewResult struct {
	Rows        []PreviewRow
	ErrorCount  int
	TotalRows   int
	BlockingErr bool
}

package audit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/scheduler"
	"sumeru/core/sdk"
)

const (
	auditModel                = "sys.audit"
	auditArchiveModel         = "sys.audit.archive"
	auditRetentionPolicyModel = "sys.audit.retention.policy"
	auditRetentionCronCode    = "audit.retention"
)

// RetentionOptions configures one retention pass.
type RetentionOptions struct {
	RetentionDays     int
	FilterModel       string
	DryRun            bool
	LegalHold         bool
	DeleteAfterExport bool
	Compress          bool
	BatchSize         int
	MaxRowsPerRun     int
}

// RetentionResult aggregates one run.
type RetentionResult struct {
	RowsExported int
	RowsDeleted  int
	RowsEligible int
	Batches      int
	Summary      string
}

func init() {
	sdk.RegisterObjectAction(auditRetentionPolicyModel, "action_run_retention", runRetentionFromPolicy)
	scheduler.RegisterCronHandler(auditRetentionCronCode, runAuditRetentionCron)
}

func runAuditRetentionCron(ctx context.Context, _ map[string]interface{}) error {
	policy, err := loadRetentionPolicy(ctx)
	if err != nil {
		return nil
	}
	if !sdk.AsBool(policy["auto_enabled"]) {
		return nil
	}
	_, err = executeRetentionPolicy(ctx, policy, false)
	return err
}

func runRetentionFromPolicy(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	if model != auditRetentionPolicyModel || 0 >= id {
		return "", fmt.Errorf("invalid policy record")
	}
	if !sdk.UserHasGroupXML(ctx, sdk.SecurityUID(ctx), "base.group_system") {
		return "", fmt.Errorf("access denied")
	}
	rec, err := sdk.SearchOne(ctx, auditRetentionPolicyModel, map[string]interface{}{"id": id})
	if err != nil {
		return "", fmt.Errorf("record not found")
	}
	dryRun := sdk.AsBool(rec["dry_run_next"])
	summary, err := executeRetentionPolicy(ctx, rec, dryRun)
	if err != nil {
		return "", err
	}
	if dryRun {
		_, _ = sdk.Write(ctx, auditRetentionPolicyModel, []int{id}, map[string]interface{}{"dry_run_next": false})
	}
	return summary, nil
}

func executeRetentionPolicy(ctx context.Context, policy map[string]interface{}, forceDryRun bool) (string, error) {
	opts := retentionOptionsFromPolicy(policy, forceDryRun)
	var result RetentionResult
	err := sdk.WithElevated(ctx, "audit_retention", func(elevCtx context.Context) error {
		var runErr error
		result, runErr = runAuditRetention(elevCtx, opts)
		return runErr
	})
	summary := result.Summary
	if summary == "" && err != nil {
		summary = err.Error()
	}
	policyID, _ := sdk.CoerceInt64(policy["id"])
	if policyID > 0 {
		_, _ = sdk.Write(ctx, auditRetentionPolicyModel, []int{int(policyID)}, map[string]interface{}{
			"last_run_at":      time.Now().UTC().Format(time.RFC3339),
			"last_run_summary": summary,
		})
	}
	if err == nil && !opts.DryRun && (result.RowsExported > 0 || result.RowsDeleted > 0) {
		sdk.AppendAudit(ctx, "audit_retention", auditModel, 0, nil, nil, summary)
	}
	return summary, err
}

func loadRetentionPolicy(ctx context.Context) (map[string]interface{}, error) {
	rows, err := sdk.SearchPage(ctx, auditRetentionPolicyModel, nil, 1, 0, "id ASC")
	if err != nil {
		return nil, err
	}
	if 0 == len(rows) {
		return nil, fmt.Errorf("audit retention policy not configured")
	}
	return rows[0], nil
}

func retentionOptionsFromPolicy(policy map[string]interface{}, forceDryRun bool) RetentionOptions {
	days, _ := sdk.CoerceInt64(policy["hot_retention_days"])
	if 0 >= days {
		if raw := strings.TrimSpace(sdk.GetConfigParam(context.Background(), "audit.default_retention_days", "")); raw != "" {
			if d, ok := sdk.CoerceInt64(raw); ok && d > 0 {
				days = d
			}
		}
	}
	if 0 >= days {
		days = 90
	}
	batch, _ := sdk.CoerceInt64(policy["batch_size"])
	maxRows, _ := sdk.CoerceInt64(policy["max_rows_per_run"])
	compress := true
	if _, ok := policy["compress"]; ok {
		compress = sdk.AsBool(policy["compress"])
	}
	return RetentionOptions{
		RetentionDays:     int(days),
		FilterModel:       strings.TrimSpace(sdk.AsString(policy["filter_model"])),
		DryRun:            forceDryRun || sdk.AsBool(policy["dry_run_next"]),
		LegalHold:         sdk.AsBool(policy["legal_hold"]),
		DeleteAfterExport: sdk.AsBool(policy["delete_after_export"]),
		Compress:          compress,
		BatchSize:         int(batch),
		MaxRowsPerRun:     int(maxRows),
	}
}

func runAuditRetention(ctx context.Context, opts RetentionOptions) (RetentionResult, error) {
	opts = normalizeRetentionOptions(opts)
	cutoff := time.Now().UTC().Add(-time.Duration(opts.RetentionDays) * 24 * time.Hour).Format(time.RFC3339)
	domain := [][]interface{}{{"create_date", "<", cutoff}}
	if fm := strings.TrimSpace(opts.FilterModel); fm != "" {
		domain = append(domain, []interface{}{"model", "=", fm})
	}

	out := RetentionResult{}
	if opts.DryRun {
		n, err := sdk.SearchCount(ctx, auditModel, domain)
		if err != nil {
			return out, err
		}
		if opts.MaxRowsPerRun > 0 && n > opts.MaxRowsPerRun {
			n = opts.MaxRowsPerRun
		}
		out.RowsEligible = n
		out.Summary = fmt.Sprintf("Dry run: %d audit row(s) older than %d days would be processed", n, opts.RetentionDays)
		return out, nil
	}

	var processed int
	for processed < opts.MaxRowsPerRun {
		limit := opts.BatchSize
		if remaining := opts.MaxRowsPerRun - processed; remaining < limit {
			limit = remaining
		}
		rows, err := sdk.SearchPage(ctx, auditModel, domain, limit, 0, "id ASC")
		if err != nil {
			return out, err
		}
		if 0 == len(rows) {
			break
		}
		if err := processAuditBatch(ctx, opts, rows, &out); err != nil {
			out.Summary = err.Error()
			return out, err
		}
		out.Batches++
		processed += len(rows)
		if len(rows) < limit {
			break
		}
	}
	out.Summary = fmt.Sprintf("exported=%d deleted=%d batches=%d", out.RowsExported, out.RowsDeleted, out.Batches)
	return out, nil
}

func normalizeRetentionOptions(opts RetentionOptions) RetentionOptions {
	if opts.RetentionDays <= 0 {
		opts.RetentionDays = 90
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = 500
	}
	if opts.MaxRowsPerRun <= 0 {
		opts.MaxRowsPerRun = 5000
	}
	return opts
}

func processAuditBatch(ctx context.Context, opts RetentionOptions, rows []map[string]interface{}, out *RetentionResult) error {
	exportRows := make([]auditExportRow, 0, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		exportRows = append(exportRows, auditRowFromMap(row))
		id, ok := sdk.CoerceInt64(row["id"])
		if ok && id > 0 {
			ids = append(ids, id)
		}
	}
	if 0 == len(exportRows) {
		return nil
	}

	raw, err := encodeAuditJSONL(exportRows)
	if err != nil {
		return err
	}
	payload, mime, err := maybeGzipAudit(opts.Compress, raw)
	if err != nil {
		return err
	}
	periodFrom, periodTo := auditPeriodBounds(exportRows)
	minID, maxID := auditMinMaxID(exportRows)
	filename := fmt.Sprintf("audit_%s_%d_%d", time.Now().UTC().Format("20060102T150405"), minID, maxID)
	if opts.Compress {
		filename += ".jsonl.gz"
	} else {
		filename += ".jsonl"
	}

	if !sdk.ModelRegistered(auditArchiveModel) {
		return fmt.Errorf("%s not registered", auditArchiveModel)
	}
	archiveID, err := sdk.Create(ctx, auditArchiveModel, map[string]interface{}{
		"name":        filename,
		"state":       auditArchiveStateStored,
		"period_from": periodFrom,
		"period_to":   periodTo,
		"row_count":   len(exportRows),
		"byte_size":   len(payload),
		"sha256":      sha256Hex(payload),
	})
	if err != nil {
		return err
	}

	attID, err := sdk.CreateBinaryAttachment(ctx, sdk.BinaryAttachmentInput{
		Name:     filename,
		ResModel: auditArchiveModel,
		ResID:    archiveID,
		Data:     payload,
		Mimetype: mime,
	})
	if err != nil {
		_, _ = sdk.Write(ctx, auditArchiveModel, []int{archiveID}, map[string]interface{}{
			"state":         auditArchiveStateFailed,
			"error_message": err.Error(),
		})
		return err
	}
	_, _ = sdk.Write(ctx, auditArchiveModel, []int{archiveID}, map[string]interface{}{
		"attachment_id": attID,
	})
	out.RowsExported += len(exportRows)

	if opts.LegalHold || !opts.DeleteAfterExport {
		return nil
	}
	if err := unlinkAuditIDs(ctx, ids); err != nil {
		return err
	}
	out.RowsDeleted += len(ids)
	_, _ = sdk.Write(ctx, auditArchiveModel, []int{archiveID}, map[string]interface{}{
		"state": auditArchiveStatePurged,
	})
	return nil
}

func unlinkAuditIDs(ctx context.Context, ids []int64) error {
	for _, id := range ids {
		if 0 >= id {
			continue
		}
		if _, err := sdk.Unlink(ctx, auditModel, []int{int(id)}); err != nil {
			return err
		}
	}
	return nil
}

// RunAuditRetentionForTest runs retention with options (tests).
func RunAuditRetentionForTest(ctx context.Context, opts RetentionOptions) (RetentionResult, error) {
	return runAuditRetention(ctx, opts)
}

// RetentionOptionsFromPolicyForTest maps policy record to options (tests).
func RetentionOptionsFromPolicyForTest(policy map[string]interface{}, dryRun bool) RetentionOptions {
	return retentionOptionsFromPolicy(policy, dryRun)
}

// RunAuditRetentionCronForTest runs cron handler (tests).
func RunAuditRetentionCronForTest(ctx context.Context) error {
	return runAuditRetentionCron(ctx, nil)
}

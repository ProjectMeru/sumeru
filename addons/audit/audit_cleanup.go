package audit

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/sdk"
)

const (
	auditCleanupModel   = "sys.audit.cleanup"
	auditPurgeBatchSize = 500
)

func init() {
	sdk.RegisterObjectAction(auditCleanupModel, "action_purge", purgeAuditLog)
}

func purgeAuditLog(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	if model != auditCleanupModel || 0 >= id {
		return "", fmt.Errorf("invalid cleanup record")
	}
	if !sdk.UserHasGroupXML(ctx, sdk.SecurityUID(ctx), "base.group_system") {
		return "", fmt.Errorf("access denied")
	}
	if err := sdk.CheckModelAccess(ctx, sdk.SecurityUID(ctx), auditModel, "unlink"); err != nil {
		return "", fmt.Errorf("access denied")
	}
	rec, err := sdk.SearchOne(ctx, auditCleanupModel, map[string]interface{}{"id": id})
	if err != nil {
		return "", fmt.Errorf("record not found")
	}
	days, err := retentionDaysFromRecord(rec)
	if err != nil {
		return "", err
	}
	dryRun := sdk.AsBool(rec["dry_run"])
	opts := RetentionOptions{
		RetentionDays:     days,
		FilterModel:       strings.TrimSpace(sdk.AsString(rec["filter_model"])),
		DryRun:            dryRun,
		DeleteAfterExport: !dryRun,
		Compress:          true,
		BatchSize:         auditPurgeBatchSize,
		MaxRowsPerRun:     50000,
	}
	if policy, perr := loadRetentionPolicy(ctx); perr == nil && len(policy) > 0 {
		policyOpts := retentionOptionsFromPolicy(policy, dryRun)
		opts.Compress = policyOpts.Compress
		opts.LegalHold = policyOpts.LegalHold
		if policyOpts.LegalHold {
			opts.DeleteAfterExport = false
		}
	}
	var result RetentionResult
	err = sdk.WithElevated(ctx, "audit_cleanup", func(elevCtx context.Context) error {
		var runErr error
		result, runErr = runAuditRetention(elevCtx, opts)
		return runErr
	})
	if err != nil {
		return "", err
	}
	summary := result.Summary
	if summary == "" {
		summary = fmt.Sprintf("%d audit row(s) older than %d days", result.RowsDeleted+result.RowsExported, days)
	}
	if dryRun {
		summary = "Dry run: " + summary
	} else if result.RowsDeleted > 0 || result.RowsExported > 0 {
		sdk.AppendAudit(ctx, "audit_cleanup", auditModel, 0, nil, nil,
			fmt.Sprintf("deleted=%d exported=%d days=%d model_filter=%q", result.RowsDeleted, result.RowsExported, days, strings.TrimSpace(sdk.AsString(rec["filter_model"]))))
	}
	_, _ = sdk.Write(ctx, auditCleanupModel, []int{id}, map[string]interface{}{"result_summary": summary})
	return "", nil
}

func retentionDaysFromRecord(rec map[string]interface{}) (int, error) {
	preset := strings.TrimSpace(sdk.AsString(rec["retention_preset"]))
	switch preset {
	case "7_days":
		return 7, nil
	case "30_days", "":
		return 30, nil
	case "custom":
		d, ok := sdk.CoerceInt64(rec["retention_days"])
		if !ok || 0 >= d {
			return 0, fmt.Errorf("custom retention requires days greater than zero")
		}
		if d > 3650 {
			return 0, fmt.Errorf("retention days too large")
		}
		return int(d), nil
	default:
		return 0, fmt.Errorf("unknown retention preset %q", preset)
	}
}

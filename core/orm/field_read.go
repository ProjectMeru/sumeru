package orm

import (
	"context"
	"fmt"
)

var sensitiveReadFields = map[string]map[string]bool{
	"core.user":        {"password": true},
	"core.user.apikey": {"key_hash": true},
}

// CheckFieldReadAccess errors if any requested field is read-denied by sys.field.access.
func CheckFieldReadAccess(ctx context.Context, uid int, model string, fields []string) error {
	if SecurityBypass(ctx) || uid == superuserUID || len(fields) == 0 {
		return nil
	}
	denied, err := fieldAccessDenied(ctx, uid, model, "read")
	if err != nil {
		return err
	}
	for _, f := range fields {
		if denied[f] {
			return fmt.Errorf("field access denied: %s.%s", model, f)
		}
	}
	return nil
}

// RedactRecordForRead removes sensitive columns and read-denied fields from a record map.
func RedactRecordForRead(ctx context.Context, uid int, model string, rec map[string]interface{}) {
	if rec == nil {
		return
	}
	for field := range sensitiveReadFields[model] {
		delete(rec, field)
	}
	if SecurityBypass(ctx) || uid == superuserUID {
		return
	}
	denied, err := fieldAccessDenied(ctx, uid, model, "read")
	if err != nil {
		return
	}
	for field := range denied {
		delete(rec, field)
	}
}

// RedactSearchResults applies read redaction to each row.
func RedactSearchResults(ctx context.Context, uid int, model string, rows []map[string]interface{}) {
	for i := range rows {
		RedactRecordForRead(ctx, uid, model, rows[i])
	}
}

func enrichRecordForRead(ctx context.Context, uid int, modelName string, record map[string]interface{}) {
	RedactRecordForRead(ctx, uid, modelName, record)
	_ = ApplyComputes(ctx, modelName, record)
	if !skipRelatedEnrichment(ctx) {
		_ = ApplyRelatedFields(ctx, modelName, record)
	}
}

func enrichRecordsForRead(ctx context.Context, uid int, modelName string, records []map[string]interface{}) {
	for _, record := range records {
		RedactRecordForRead(ctx, uid, modelName, record)
		_ = ApplyComputes(ctx, modelName, record)
	}
	if !skipRelatedEnrichment(ctx) {
		_ = ApplyRelatedFieldsBatch(ctx, modelName, records)
	}
}

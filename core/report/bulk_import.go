package report

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// ExecuteBulkImportInput runs import after user confirms mapping.
type ExecuteBulkImportInput struct {
	BatchID       int
	Mapping       map[string]string
	SkipInvalid   bool
	ImportValidOnly bool
}

// ExecuteBulkImport creates or updates records from a staged batch.
func ExecuteBulkImport(ctx context.Context, in ExecuteBulkImportInput) (ImportResult, error) {
	batch, data, storedMapping, err := loadBatchCSV(ctx, in.BatchID)
	if err != nil {
		return ImportResult{}, err
	}
	targetModel := orm.AsString(batch["target_model"])
	mode := orm.AsString(batch["import_mode"])
	modelInst, ok := orm.Registry[targetModel]
	if !ok {
		return ImportResult{}, fmt.Errorf("unknown model %q", targetModel)
	}
	mapping := in.Mapping
	if len(mapping) == 0 {
		mapping = storedMapping
	}
	headers, rows, err := parseCSV(data)
	if err != nil {
		return ImportResult{}, err
	}
	allowed := allowedFieldNames(modelInst)
	var result ImportResult
	for i, record := range rows {
		vals := rowValuesFromMapping(headers, record, mapping)
		if len(vals) == 0 {
			result.Skipped++
			continue
		}
		errs := validateRowValues(modelInst, vals, allowed, mode)
		if len(errs) > 0 {
			if in.SkipInvalid || in.ImportValidOnly {
				result.Skipped++
				continue
			}
			return result, fmt.Errorf("row %d: %s", i+1, strings.Join(errs, "; "))
		}
		if mode == ImportModeUpsert {
			if idRaw, ok := vals["id"]; ok {
				if id, ok := orm.CoerceInt64(idRaw); ok && id > 0 {
					updateVals := map[string]interface{}{}
					for k, v := range vals {
						if k != "id" {
							updateVals[k] = v
						}
					}
					if len(updateVals) > 0 {
						if err := orm.UpdateRecordByID(ctx, targetModel, int(id), updateVals); err != nil {
							if in.SkipInvalid {
								result.Skipped++
								continue
							}
							return result, fmt.Errorf("row %d update: %w", i+1, err)
						}
						result.Updated++
						continue
					}
				}
			}
		}
		delete(vals, "id")
		if _, err := orm.Create(ctx, modelInst, vals); err != nil {
			if in.SkipInvalid {
				result.Skipped++
				continue
			}
			return result, fmt.Errorf("row %d create: %w", i+1, err)
		}
		result.Created++
	}
	if err := orm.UpdateRecordByID(ctx, BulkModelName, in.BatchID, map[string]interface{}{"state": "done"}); err != nil {
		return result, fmt.Errorf("mark batch done: %w", err)
	}
	return result, nil
}

// CancelBatch marks batch cancelled.
func CancelBatch(ctx context.Context, batchID int) error {
	return orm.UpdateRecordByID(ctx, BulkModelName, batchID, map[string]interface{}{"state": "cancelled"})
}

// ImportFlashMessage formats redirect flash text.
func ImportFlashMessage(r ImportResult) string {
	return fmt.Sprintf("imported_%d_updated_%d_skipped_%d", r.Created, r.Updated, r.Skipped)
}

// ParseMappingJSON parses column mapping from form/json.
func ParseMappingJSON(raw string) (map[string]string, error) {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

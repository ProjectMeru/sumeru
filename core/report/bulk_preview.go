package report

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// PreviewBulkImportInput validates mapped rows before confirm.
type PreviewBulkImportInput struct {
	BatchID int
	Mapping map[string]string
}

// PreviewBulkImport validates up to 10 rows with the given column mapping.
func PreviewBulkImport(ctx context.Context, in PreviewBulkImportInput) (PreviewResult, error) {
	batch, data, _, err := loadBatchCSV(ctx, in.BatchID)
	if err != nil {
		return PreviewResult{}, err
	}
	targetModel := orm.AsString(batch["target_model"])
	modelInst, ok := orm.Registry[targetModel]
	if !ok {
		return PreviewResult{}, fmt.Errorf("unknown model %q", targetModel)
	}
	headers, rows, err := parseCSV(data)
	if err != nil {
		return PreviewResult{}, err
	}
	mapping := in.Mapping
	if 0 == len(mapping) {
		_ = json.Unmarshal([]byte(orm.AsString(batch["column_mapping"])), &mapping)
	}
	allowed := allowedFieldNames(modelInst)
	result := PreviewResult{TotalRows: len(rows)}
	limit := maxPreviewRows
	if len(rows) < limit {
		limit = len(rows)
	}
	for i := 0; i < limit; i++ {
		vals := rowValuesFromMapping(headers, rows[i], mapping)
		coerced, coerceErrs := coerceImportValues(ctx, modelInst, vals)
		vals = coerced
		var errs []string
		errs = append(errs, coerceErrs...)
		errs = append(errs, validateRowValues(modelInst, vals, allowed, orm.AsString(batch["import_mode"]))...)
		if len(errs) > 0 {
			result.ErrorCount += len(errs)
			result.BlockingErr = true
		}
		result.Rows = append(result.Rows, PreviewRow{
			RowNum: i + 1,
			Values: vals,
			Errors: errs,
		})
	}
	return result, nil
}

func validateRowValues(modelInst orm.Model, vals map[string]interface{}, allowed map[string]struct{}, mode string) []string {
	var errs []string
	if 0 == len(vals) {
		return []string{"no mapped values"}
	}
	for k := range vals {
		if _, ok := allowed[k]; !ok {
			errs = append(errs, fmt.Sprintf("unknown field %q", k))
		}
	}
	if mode == ImportModeUpsert {
		if idRaw, ok := vals["id"]; ok {
			if id, ok := orm.CoerceInt64(idRaw); !ok || 0 >= id {
				errs = append(errs, "invalid id for upsert")
			}
		}
	}
	for _, f := range modelInst.Fields() {
		if !f.Required {
			continue
		}
		if f.Name == "id" {
			continue
		}
		if _, ok := vals[f.Name]; !ok {
			if mode == ImportModeUpsert {
				continue
			}
			errs = append(errs, fmt.Sprintf("missing required field %q", f.Name))
		} else if strings.TrimSpace(orm.AsString(vals[f.Name])) == "" {
			errs = append(errs, fmt.Sprintf("empty required field %q", f.Name))
		}
	}
	return errs
}

// UpdateBatchMapping saves column mapping JSON on the batch row.
func UpdateBatchMapping(ctx context.Context, batchID int, mapping map[string]string) error {
	raw, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	return orm.UpdateRecordByID(ctx, BulkModelName, batchID, map[string]interface{}{
		"column_mapping": string(raw),
	})
}

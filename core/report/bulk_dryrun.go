package report

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

// ImportMessage is one row/column validation or simulation message.
type ImportMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Row     int    `json:"row,omitempty"`
}

// DryRunBulkImport validates every row and simulates create/update counts without writes.
func DryRunBulkImport(ctx context.Context, in PreviewBulkImportInput) (PreviewResult, []ImportMessage, error) {
	batch, data, _, err := loadBatchCSV(ctx, in.BatchID)
	if err != nil {
		return PreviewResult{}, nil, err
	}
	targetModel := orm.AsString(batch["target_model"])
	modelInst, ok := orm.Registry[targetModel]
	if !ok {
		return PreviewResult{}, nil, fmt.Errorf("unknown model %q", targetModel)
	}
	headers, rows, err := parseCSV(data)
	if err != nil {
		return PreviewResult{}, nil, err
	}
	mapping := in.Mapping
	if 0 == len(mapping) {
		_ = json.Unmarshal([]byte(orm.AsString(batch["column_mapping"])), &mapping)
	}
	mode := orm.AsString(batch["import_mode"])
	allowed := allowedFieldNames(modelInst)
	result := PreviewResult{TotalRows: len(rows)}
	var messages []ImportMessage
	var wouldCreate, wouldUpdate int

	for i, record := range rows {
		vals := rowValuesFromMapping(headers, record, mapping)
		if 0 == len(vals) {
			continue
		}
		coerced, coerceErrs := coerceImportValues(ctx, modelInst, vals)
		for _, e := range coerceErrs {
			messages = append(messages, ImportMessage{Type: "error", Message: e, Row: i + 1})
			result.ErrorCount++
		}
		vals = coerced
		errs := validateRowValues(modelInst, vals, allowed, mode)
		for _, e := range errs {
			messages = append(messages, ImportMessage{Type: "error", Message: e, Row: i + 1})
			result.ErrorCount++
		}
		if len(errs) > 0 || len(coerceErrs) > 0 {
			result.BlockingErr = true
			continue
		}
		if mode == ImportModeUpsert {
			if idRaw, ok := vals["id"]; ok {
				if id, ok := orm.CoerceInt64(idRaw); ok && id > 0 {
					if _, err := orm.SearchOne(ctx, targetModel, map[string]interface{}{"id": int(id)}); err == nil {
						wouldUpdate++
						continue
					}
				}
			}
		}
		wouldCreate++
	}
	result.Rows = nil
	summary := fmt.Sprintf("Would create %d, update %d", wouldCreate, wouldUpdate)
	messages = append([]ImportMessage{{Type: "info", Message: summary}}, messages...)
	return result, messages, nil
}

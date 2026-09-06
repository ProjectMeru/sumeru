package report

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sumeru/core/orm"
)

// CreateBatchInput stages an uploaded CSV for mapping.
type CreateBatchInput struct {
	TargetModel    string
	ImportMode     string
	SelectedFields []string
	CSVContent     []byte
	NextURL        string
	UserID         int
	ActionID       int
}

// CreateBatch stores CSV in sys.attachment and creates sys.bulk.import row.
func CreateBatch(ctx context.Context, in CreateBatchInput) (batchID int, err error) {
	if in.TargetModel == "" {
		return 0, fmt.Errorf("target model required")
	}
	mode := strings.ToLower(strings.TrimSpace(in.ImportMode))
	if mode != ImportModeCreate && mode != ImportModeUpsert {
		mode = ImportModeCreate
	}
	headers, _, err := parseCSV(in.CSVContent)
	if err != nil {
		return 0, err
	}
	fieldsJSON, _ := json.Marshal(in.SelectedFields)
	headersJSON, _ := json.Marshal(headers)

	attID, err := orm.Create(ctx, orm.Registry["sys.attachment"], map[string]interface{}{
		"name":      fmt.Sprintf("bulk_%s.csv", time.Now().Format("20060102_150405")),
		"model":     BulkModelName,
		"mimetype":  "text/csv",
		"file_size": len(in.CSVContent),
		"datas":     base64.StdEncoding.EncodeToString(in.CSVContent),
	})
	if err != nil {
		return 0, fmt.Errorf("stage attachment: %w", err)
	}

	mapping := defaultColumnMapping(headers, in.SelectedFields)
	mappingJSON, _ := json.Marshal(mapping)

	batchID, err = orm.Create(ctx, orm.Registry[BulkModelName], map[string]interface{}{
		"name":             fmt.Sprintf("Import %s", in.TargetModel),
		"target_model":     in.TargetModel,
		"import_mode":      mode,
		"attachment_id":    attID,
		"selected_fields":  string(fieldsJSON),
		"csv_headers":      string(headersJSON),
		"column_mapping":   string(mappingJSON),
		"next_url":         in.NextURL,
		"user_id":          in.UserID,
		"action_id":        in.ActionID,
		"state":            "draft",
	})
	if err != nil {
		return 0, err
	}
	if err := orm.UpdateRecordByID(ctx, "sys.attachment", attID, map[string]interface{}{"res_id": batchID}); err != nil {
		return batchID, fmt.Errorf("link attachment to batch: %w", err)
	}
	return batchID, nil
}

func defaultColumnMapping(headers, selectedFields []string) map[string]string {
	selected := map[string]struct{}{}
	for _, f := range selectedFields {
		selected[f] = struct{}{}
	}
	mapping := map[string]string{}
	for _, h := range headers {
		hLower := strings.ToLower(strings.TrimSpace(h))
		if _, ok := selected[hLower]; ok {
			mapping[h] = hLower
			continue
		}
		if _, ok := selected[h]; ok {
			mapping[h] = h
			continue
		}
		mapping[h] = ""
	}
	return mapping
}

func loadBatchCSV(ctx context.Context, batchID int) (map[string]interface{}, []byte, map[string]string, error) {
	batch, err := orm.SearchOne(ctx, BulkModelName, map[string]interface{}{"id": batchID})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("batch not found")
	}
	attID, _ := orm.CoerceInt64(batch["attachment_id"])
	if attID <= 0 {
		return nil, nil, nil, fmt.Errorf("missing attachment")
	}
	att, err := orm.SearchOne(ctx, "sys.attachment", map[string]interface{}{"id": attID})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("attachment not found")
	}
	raw := orm.AsString(att["datas"])
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		data = []byte(raw)
	}
	mapping := map[string]string{}
	_ = json.Unmarshal([]byte(orm.AsString(batch["column_mapping"])), &mapping)
	return batch, data, mapping, nil
}

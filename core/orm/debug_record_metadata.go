package orm

import (
	"context"
	"fmt"
	"strings"
)

const debugMetadataXmlIDCap = 50

// BuildDebugRecordMetadata returns record audit + XML id rows for developer UI (admin support only).
func BuildDebugRecordMetadata(ctx context.Context, uid int, model string, recordID int) (map[string]interface{}, error) {
	model = strings.TrimSpace(model)
	if 0 >= uid {
		return nil, fmt.Errorf("login required")
	}
	if UserIsPortalOnly(ctx, uid) {
		return nil, fmt.Errorf("access denied")
	}
	if model == "" {
		return nil, fmt.Errorf("model required")
	}
	if 0 >= recordID {
		return nil, fmt.Errorf("record_id required")
	}
	if RegistryModel(model) == nil {
		return nil, fmt.Errorf("unknown model %q", model)
	}
	if err := CheckModelAccess(ctx, uid, model, "read"); err != nil {
		return nil, err
	}

	rows, err := SearchLimit(ctx, model, [][]interface{}{[]interface{}{"id", "=", int64(recordID)}}, 1)
	if err != nil {
		return nil, err
	}
	if 0 == len(rows) {
		return nil, fmt.Errorf("record not found")
	}
	rec := rows[0]

	out := map[string]interface{}{
		"id": recordID,
	}
	for _, key := range []string{"create_uid", "write_uid", "create_date", "write_date"} {
		if v, ok := rec[key]; ok && v != nil {
			out[key] = v
		}
	}

	xmlRows, err := SearchLimit(ctx, "sys.model.data", [][]interface{}{
		{"model", "=", model},
		{"core_id", "=", int64(recordID)},
	}, debugMetadataXmlIDCap)
	if err != nil {
		return nil, err
	}
	xmlIDs := make([]map[string]interface{}, 0, len(xmlRows))
	for _, row := range xmlRows {
		mod := strings.TrimSpace(AsString(row["module"]))
		name := strings.TrimSpace(AsString(row["name"]))
		if mod == "" || name == "" {
			continue
		}
		xmlIDs = append(xmlIDs, map[string]interface{}{
			"xml_id": mod + "." + name,
		})
	}
	out["xml_ids"] = xmlIDs
	return out, nil
}

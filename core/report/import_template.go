package report

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

const importTemplateModel = "sys.import.template"

// ApplyImportTemplateMapping merges saved template mapping when headers match.
func ApplyImportTemplateMapping(ctx context.Context, targetModel string, headers []string, uid int) map[string]string {
	targetModel = strings.TrimSpace(targetModel)
	if targetModel == "" {
		return nil
	}
	domain := [][]interface{}{{"target_model", "=", targetModel}}
	if uid > 0 {
		domain = append(domain, []interface{}{"user_id", "=", uid})
	}
	rows, err := orm.SearchPage(ctx, importTemplateModel, domain, 5, 0, "id DESC")
	if err != nil || 0 == len(rows) {
		domain = [][]interface{}{{"target_model", "=", targetModel}, {"shared", "=", true}}
		rows, err = orm.SearchPage(ctx, importTemplateModel, domain, 5, 0, "id DESC")
	}
	if err != nil {
		return nil
	}
	if 0 == len(rows) {
		return nil
	}

	var mapping map[string]string
	_ = json.Unmarshal([]byte(orm.AsString(rows[0]["column_mapping"])), &mapping)
	if 0 == len(mapping) {
		return nil
	}
	out := map[string]string{}
	for _, h := range headers {
		if f, ok := mapping[h]; ok && strings.TrimSpace(f) != "" {
			out[h] = f
		}
	}
	return out
}

// SaveImportTemplate persists mapping for reuse.
func SaveImportTemplate(ctx context.Context, name, targetModel, importMode, selectedFields, columnMapping string, uid int, shared bool) (int, error) {
	inst, ok := orm.Registry[importTemplateModel]
	if !ok {
		return 0, fmt.Errorf("import template model missing")
	}
	vals := map[string]interface{}{
		"name":            strings.TrimSpace(name),
		"target_model":    targetModel,
		"import_mode":     importMode,
		"selected_fields": selectedFields,
		"column_mapping":  columnMapping,
		"shared":          shared,
	}
	if uid > 0 {
		vals["user_id"] = uid
	}
	return orm.Create(ctx, inst, vals)
}

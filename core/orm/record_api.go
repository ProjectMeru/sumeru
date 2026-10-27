package orm

import (
	"context"
	"fmt"
	"strings"
)

const nameSearchDefaultLimit = 8

// NameSearch returns id and display label rows for many2one completion.
// An empty query returns the first page. Models without a name or login field return no rows.
func NameSearch(ctx context.Context, modelName, query string, limit int) ([]map[string]interface{}, error) {
	modelName = strings.TrimSpace(modelName)
	if _, err := requireRegisteredModel(modelName); err != nil {
		return nil, err
	}
	var searchable []string
	for _, field := range []string{"name", "login"} {
		if fieldSearchable(modelName, field) {
			searchable = append(searchable, field)
		}
	}
	if 0 == len(searchable) {
		return []map[string]interface{}{}, nil
	}
	if 0 >= limit {
		limit = nameSearchDefaultLimit
	}
	domain := BuildListSearchDomain(modelName, searchable, query)
	rows, err := SearchPage(ctx, modelName, domain, limit, 0, "id ASC")
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		id, _ := CoerceInt64(row["id"])
		label := strings.TrimSpace(AsString(row["name"]))
		if label == "" {
			label = strings.TrimSpace(AsString(row["login"]))
		}
		if label == "" {
			label = fmt.Sprintf("%d", id)
		}
		out = append(out, map[string]interface{}{"id": int(id), "name": label})
	}
	return out, nil
}

// NameCreate inserts a row whose name is the completion text.
func NameCreate(ctx context.Context, modelName, name string) (int, error) {
	modelName = strings.TrimSpace(modelName)
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("name required")
	}
	inst, err := requireRegisteredModel(modelName)
	if err != nil {
		return 0, err
	}
	if FieldDef(modelName, "name") == nil {
		return 0, fmt.Errorf("model %s has no name field", modelName)
	}
	return Create(ctx, inst, map[string]interface{}{"name": name})
}

// DefaultGet builds a new-row map from field defaults, then applies overlay.
// overlay wins. fields limits the returned keys; an empty list returns every default.
func DefaultGet(ctx context.Context, modelName string, fields []string, overlay map[string]interface{}) (map[string]interface{}, error) {
	modelName = strings.TrimSpace(modelName)
	inst, err := requireRegisteredModel(modelName)
	if err != nil {
		return nil, err
	}
	if err := RejectSmuggledUserBypass(ctx); err != nil {
		return nil, err
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "create"); err != nil {
		return nil, err
	}
	want := map[string]struct{}{}
	for _, name := range fields {
		name = strings.TrimSpace(name)
		if name == "" || name == "id" {
			continue
		}
		if FieldDef(modelName, name) == nil {
			return nil, fmt.Errorf("unknown field %q on model %s", name, modelName)
		}
		want[name] = struct{}{}
	}
	out := map[string]interface{}{}
	defs := fieldDefinitionsByName(inst)
	if err := applySpecialDefaults(ctx, defs, out); err != nil {
		return nil, err
	}
	for name, field := range defs {
		if _, ok := out[name]; ok {
			continue
		}
		if field.DefaultVal == nil || isRuntimeDefaultToken(field.DefaultVal) {
			continue
		}
		if IsVirtualField(field) || field.Type == Many2Many || field.Type == One2Many {
			continue
		}
		out[name] = field.DefaultVal
	}
	for key, value := range overlay {
		key = strings.TrimSpace(key)
		if key == "" || FieldDef(modelName, key) == nil {
			continue
		}
		out[key] = value
	}
	if 0 == len(want) {
		return out, nil
	}
	filtered := map[string]interface{}{}
	for key := range want {
		if value, ok := out[key]; ok {
			filtered[key] = value
		}
	}
	return filtered, nil
}

// Copy duplicates one row through Create. id, unique columns, and audit stamps are omitted.
func Copy(ctx context.Context, modelName string, id int) (int, error) {
	modelName = strings.TrimSpace(modelName)
	if 0 >= id {
		return 0, fmt.Errorf("invalid id")
	}
	inst, err := requireRegisteredModel(modelName)
	if err != nil {
		return 0, err
	}
	rec, err := SearchOne(ctx, modelName, map[string]interface{}{"id": id})
	if err != nil {
		return 0, err
	}
	vals := map[string]interface{}{}
	for _, field := range inst.Fields() {
		if skipCopyField(field) {
			continue
		}
		value, ok := rec[field.Name]
		if !ok || value == nil {
			continue
		}
		vals[field.Name] = value
	}
	return Create(ctx, inst, vals)
}

func skipCopyField(field FieldDefinition) bool {
	switch field.Name {
	case "", "id", "create_uid", "write_uid", "create_date", "write_date":
		return true
	}
	if field.Unique || IsVirtualField(field) {
		return true
	}
	return field.Type == Many2Many || field.Type == One2Many
}

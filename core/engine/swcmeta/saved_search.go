package swcmeta

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// LoadSavedSearches returns saved searches for the current user and model.
func LoadSavedSearches(ctx context.Context, actionID int, model string) []SavedSearchMeta {
	model = strings.TrimSpace(model)
	if model == "" || orm.DB == nil {
		return nil
	}
	uid := orm.UIDFromContext(ctx)
	if uid <= 0 {
		return nil
	}
	domain := [][]interface{}{
		{"user_id", "=", int64(uid)},
		{"model", "=", model},
	}
	if actionID > 0 {
		domain = append(domain, []interface{}{"action_id", "=", int64(actionID)})
	}
	rows, err := orm.SearchLimit(ctx, "swc.saved.search", domain, 50)
	if err != nil {
		return nil
	}
	out := make([]SavedSearchMeta, 0, len(rows))
	for _, row := range rows {
		id, _ := orm.CoerceInt64(row["id"])
		if id <= 0 {
			continue
		}
		out = append(out, SavedSearchMeta{
			ID:        int(id),
			Name:      strings.TrimSpace(orm.AsString(row["name"])),
			Search:    strings.TrimSpace(orm.AsString(row["search_query"])),
			Filter:    strings.TrimSpace(orm.AsString(row["filter_csv"])),
			Domain:    strings.TrimSpace(orm.AsString(row["domain_json"])),
			GroupBy:   strings.TrimSpace(orm.AsString(row["group_by"])),
			IsDefault: orm.AsBool(row["is_default"]),
		})
	}
	return out
}

// SaveSavedSearch creates a saved search for the current user.
func SaveSavedSearch(ctx context.Context, actionID int, model, name, search, filter, domain, groupBy string, isDefault bool) (int, error) {
	model = strings.TrimSpace(model)
	name = strings.TrimSpace(name)
	if model == "" || name == "" {
		return 0, fmt.Errorf("model and name required")
	}
	uid := orm.UIDFromContext(ctx)
	if uid <= 0 {
		return 0, fmt.Errorf("login required")
	}
	vals := map[string]interface{}{
		"name":          name,
		"user_id":       int64(uid),
		"model":         model,
		"action_id":     int64(actionID),
		"search_query":  strings.TrimSpace(search),
		"filter_csv":    strings.TrimSpace(filter),
		"domain_json":   strings.TrimSpace(domain),
		"group_by":      strings.TrimSpace(groupBy),
		"is_default":    isDefault,
	}
	return orm.Create(ctx, orm.RegistryModel("swc.saved.search"), vals)
}

// DeleteSavedSearch removes a saved search owned by the current user.
func DeleteSavedSearch(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	uid := orm.UIDFromContext(ctx)
	if uid <= 0 {
		return fmt.Errorf("login required")
	}
	row, err := orm.SearchOne(ctx, "swc.saved.search", map[string]interface{}{"id": id})
	if err != nil {
		return err
	}
	owner, _ := orm.CoerceInt64(row["user_id"])
	if int(owner) != uid {
		return fmt.Errorf("access denied")
	}
	return orm.Unlink(ctx, "swc.saved.search", id)
}

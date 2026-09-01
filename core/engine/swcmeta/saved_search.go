package swcmeta

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// LoadSavedSearches returns saved searches for the current user and model (including shared).
func LoadSavedSearches(ctx context.Context, actionID int, model string) []SavedSearchMeta {
	model = strings.TrimSpace(model)
	if model == "" || orm.DB == nil {
		return nil
	}
	uid := orm.UIDFromContext(ctx)
	if uid <= 0 {
		return nil
	}
	seen := map[int]struct{}{}
	var out []SavedSearchMeta
	appendRows := func(domain [][]interface{}) {
		if actionID > 0 {
			domain = append(domain, []interface{}{"action_id", "=", int64(actionID)})
		}
		rows, err := orm.SearchLimit(ctx, "swc.saved.search", domain, 50)
		if err != nil {
			return
		}
		for _, row := range rows {
			id, _ := orm.CoerceInt64(row["id"])
			if id <= 0 {
				continue
			}
			if _, ok := seen[int(id)]; ok {
				continue
			}
			seen[int(id)] = struct{}{}
			out = append(out, SavedSearchMeta{
				ID:        int(id),
				Name:      strings.TrimSpace(orm.AsString(row["name"])),
				Search:    strings.TrimSpace(orm.AsString(row["search_query"])),
				Filter:    strings.TrimSpace(orm.AsString(row["filter_csv"])),
				Domain:    strings.TrimSpace(orm.AsString(row["domain_json"])),
				GroupBy:   strings.TrimSpace(orm.AsString(row["group_by"])),
				IsDefault: orm.AsBool(row["is_default"]),
				IsShared:  orm.AsBool(row["is_shared"]),
			})
		}
	}
	appendRows([][]interface{}{
		{"user_id", "=", int64(uid)},
		{"model", "=", model},
	})
	appendRows([][]interface{}{
		{"is_shared", "=", true},
		{"model", "=", model},
	})
	return out
}

// SavedSearchInput holds fields for creating a saved search.
type SavedSearchInput struct {
	ActionID  int
	Model     string
	Name      string
	Search    string
	Filter    string
	Domain    string
	GroupBy   string
	IsDefault bool
	IsShared  bool
}

// SaveSavedSearch creates a saved search for the current user.
func SaveSavedSearch(ctx context.Context, in SavedSearchInput) (int, error) {
	model := strings.TrimSpace(in.Model)
	name := strings.TrimSpace(in.Name)
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
		"action_id":     int64(in.ActionID),
		"search_query":  strings.TrimSpace(in.Search),
		"filter_csv":    strings.TrimSpace(in.Filter),
		"domain_json":   strings.TrimSpace(in.Domain),
		"group_by":      strings.TrimSpace(in.GroupBy),
		"is_default":    in.IsDefault,
		"is_shared":     in.IsShared,
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
	shared := orm.AsBool(row["is_shared"])
	if int(owner) != uid && !shared {
		return fmt.Errorf("access denied")
	}
	if shared && int(owner) != uid {
		return fmt.Errorf("cannot delete shared favorite owned by another user")
	}
	return orm.Unlink(ctx, "swc.saved.search", id)
}

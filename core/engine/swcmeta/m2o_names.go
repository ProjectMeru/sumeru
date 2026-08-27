package swcmeta

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

// enrichMany2OneNames resolves display names for Many2One fields and stores them
// as `<field>_name` on each row, so the client can show e.g. the customer name
// instead of "#<id>". It batch-resolves each comodel once.
func enrichMany2OneNames(ctx context.Context, model string, rows []map[string]interface{}) {
	inst, ok := orm.Registry[model]
	if !ok || inst == nil || len(rows) == 0 {
		return
	}
	for _, fd := range inst.Fields() {
		if fd.Type != orm.Many2One || strings.TrimSpace(fd.Relation) == "" {
			continue
		}
		fieldName := fd.Name
		comodel := strings.TrimSpace(fd.Relation)
		ids := collectDistinctIDs(rows, fieldName)
		if len(ids) == 0 {
			continue
		}
		values := make([]interface{}, 0, len(ids))
		for _, id := range ids {
			values = append(values, id)
		}
		refs, err := orm.Search(ctx, comodel, [][]interface{}{{"id", "in", values}})
		if err != nil {
			continue
		}
		nameByID := make(map[int64]string, len(refs))
		for _, ref := range refs {
			id, ok := orm.CoerceInt64(ref["id"])
			if !ok {
				continue
			}
			nameByID[id] = recordDisplayName(ref)
		}
		for _, row := range rows {
			if id, ok := orm.CoerceInt64(row[fieldName]); ok {
				if name, ok := nameByID[id]; ok {
					row[fieldName+"_name"] = name
				}
			}
		}
	}
}

// EnrichMany2OneNames resolves `<field>_name` display names for Many2One fields
// on the given rows (exported for reuse by the read RPC).
func EnrichMany2OneNames(ctx context.Context, model string, rows []map[string]interface{}) {
	enrichMany2OneNames(ctx, model, rows)
}

func collectDistinctIDs(rows []map[string]interface{}, field string) []int64 {
	seen := make(map[int64]struct{})
	out := make([]int64, 0, 4)
	for _, row := range rows {
		if id, ok := orm.CoerceInt64(row[field]); ok && id > 0 {
			if _, dup := seen[id]; !dup {
				seen[id] = struct{}{}
				out = append(out, id)
			}
		}
	}
	return out
}

// recordDisplayName mirrors orm.DisplayNameForID: prefer name, then login.
func recordDisplayName(rec map[string]interface{}) string {
	if n := strings.TrimSpace(orm.AsString(rec["name"])); n != "" {
		return n
	}
	if n := strings.TrimSpace(orm.AsString(rec["login"])); n != "" {
		return n
	}
	return ""
}

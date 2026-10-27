package orm

import (
	"context"
	"fmt"
	"strings"
)

// ResolveInverseOne2ManyField finds the Many2One on comodel pointing at parentModel.
func ResolveInverseOne2ManyField(parentModel, comodel string) string {
	m, ok := Registry[comodel]
	if !ok || m == nil {
		return ""
	}
	var fallback string
	for _, f := range m.Fields() {
		if f.Type != Many2One || f.Relation != parentModel {
			continue
		}
		if strings.HasSuffix(f.Name, "_id") {
			return f.Name
		}
		if fallback == "" {
			fallback = f.Name
		}
	}
	return fallback
}

// DisplayNameForID returns a short label for a related record id.
func DisplayNameForID(ctx context.Context, modelName string, id int) string {
	if 0 >= id || strings.TrimSpace(modelName) == "" {
		return ""
	}
	rec, err := SearchOne(ctx, modelName, map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Sprintf("%d", id)
	}
	if n := strings.TrimSpace(AsString(rec["name"])); n != "" {
		return n
	}
	if n := strings.TrimSpace(AsString(rec["login"])); n != "" {
		return n
	}
	return fmt.Sprintf("%d", id)
}

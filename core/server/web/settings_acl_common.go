package web

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

const matrixMaxGroupsDefault = 40

type aclGroupCol struct {
	ID    int
	Label string
}

func resolveSettingsMenuXMLID(ctx context.Context, xmlID, fallbackMenuID string) string {
	if id := render.MenuIDForXMLID(ctx, xmlID); id != "" {
		return id
	}
	return strings.TrimSpace(fallbackMenuID)
}

func loadACLGroups(ctx context.Context, groupFilter string) ([]aclGroupCol, error) {
	rows, err := orm.Search(ctx, "core.group", nil)
	if err != nil {
		return nil, err
	}
	groupFilter = strings.ToLower(strings.TrimSpace(groupFilter))
	out := make([]aclGroupCol, 0, len(rows))
	for _, row := range rows {
		id := intField(row["id"])
		if 0 >= id {
			continue
		}
		label := stringField(row["name"])
		if groupFilter != "" && !strings.Contains(strings.ToLower(label), groupFilter) {
			continue
		}
		out = append(out, aclGroupCol{ID: id, Label: label})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out, nil
}

func listRegistryModelNames() []string {
	names := make([]string, 0, len(orm.Registry))
	for name := range orm.Registry {
		if strings.TrimSpace(name) != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func matrixGroupsForDisplay(all []aclGroupCol, showAll bool) ([]aclGroupCol, bool) {
	const globalLabel = "Global"
	cols := append([]aclGroupCol{{ID: 0, Label: globalLabel}}, all...)
	if showAll || len(cols) <= matrixMaxGroupsDefault {
		return cols, false
	}
	return cols[:matrixMaxGroupsDefault], true
}

func formHas(form map[string][]string, key string) bool {
	vals, ok := form[key]
	return ok && len(vals) > 0 && strings.TrimSpace(vals[0]) != ""
}

func matrixCellKey(groupID int, field string) string {
	return fmt.Sprintf("%d_%s", groupID, field)
}

func stringField(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func intField(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(t))
		return n
	default:
		return 0
	}
}

func boolField(v interface{}, defaultVal bool) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(strings.TrimSpace(t), "true") || strings.TrimSpace(t) == "1"
	default:
		return defaultVal
	}
}

func filterFieldNames(fields []string, fieldFilter string) []string {
	fieldFilter = strings.ToLower(strings.TrimSpace(fieldFilter))
	if fieldFilter == "" {
		return fields
	}
	out := fields[:0]
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), fieldFilter) {
			out = append(out, f)
		}
	}
	return out
}

func writeCSVAttachment(w http.ResponseWriter, filename, body string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", safeContentDispositionFilename(filename))
	_, _ = w.Write([]byte(body))
}

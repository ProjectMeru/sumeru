package web

import (
	"fmt"
	"sort"
	"strings"

	"sumeru/core/engine/render"
)

const emptyGroupKey = "__empty__"

func groupRowKey(v interface{}) string {
	if v == nil {
		return emptyGroupKey
	}
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) == "" {
			return emptyGroupKey
		}
		return t
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" || s == "<nil>" {
			return emptyGroupKey
		}
		return s
	}
}

func groupSectionLabel(key string, displayFromRow string) string {
	if displayFromRow != "" && displayFromRow != "<nil>" {
		return displayFromRow
	}
	if key == emptyGroupKey {
		return "(Empty)"
	}
	return key
}

// partitionListSections groups flat list rows by the first group-by field value.
func partitionListSections(rows []map[string]interface{}, groupField string) []render.ListSection {
	if groupField == "" || 0 == len(rows) {
		return nil
	}
	buckets := map[string][]map[string]interface{}{}
	order := []string{}
	for _, row := range rows {
		key := groupRowKey(row[groupField])
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], row)
	}
	sort.Strings(order)
	out := make([]render.ListSection, 0, len(order))
	for _, key := range order {
		sectionRows := buckets[key]
		displayFromRow := ""
		if len(sectionRows) > 0 {
			if n, ok := sectionRows[0][groupField+"_name"]; ok {
				displayFromRow = strings.TrimSpace(fmt.Sprint(n))
			}
		}
		label := groupSectionLabel(key, displayFromRow)
		out = append(out, render.ListSection{
			Label: label,
			Value: key,
			Count: len(sectionRows),
			Rows:  sectionRows,
		})
	}
	return out
}

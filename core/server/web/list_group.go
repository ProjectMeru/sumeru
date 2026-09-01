package web

import (
	"fmt"
	"sort"

	"sumeru/core/engine/render"
)

// partitionListSections groups flat list rows by the first group-by field value.
func partitionListSections(rows []map[string]interface{}, groupField string) []render.ListSection {
	if groupField == "" || len(rows) == 0 {
		return nil
	}
	buckets := map[string][]map[string]interface{}{}
	order := []string{}
	for _, row := range rows {
		key := fmt.Sprint(row[groupField])
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], row)
	}
	sort.Strings(order)
	out := make([]render.ListSection, 0, len(order))
	for _, key := range order {
		label := key
		sectionRows := buckets[key]
		if len(sectionRows) > 0 {
			if n, ok := sectionRows[0][groupField+"_name"]; ok && fmt.Sprint(n) != "" {
				label = fmt.Sprint(n)
			}
		}
		out = append(out, render.ListSection{
			Label: label,
			Value: key,
			Count: len(sectionRows),
			Rows:  sectionRows,
		})
	}
	return out
}

package swcmeta

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// BuildPivotData aggregates list rows for a pivot view definition.
func BuildPivotData(view *parser.View, rows []map[string]interface{}) *PivotMeta {
	if view == nil || 0 == len(view.Field) {
		return nil
	}
	var rowFields, colFields, measureFields []parser.Field
	for _, f := range view.Field {
		switch strings.ToLower(strings.TrimSpace(f.PivotType)) {
		case "row":
			rowFields = append(rowFields, f)
		case "col", "column":
			colFields = append(colFields, f)
		case "measure":
			measureFields = append(measureFields, f)
		}
	}
	if 0 == len(measureFields) {
		return nil
	}
	measure := measureFields[0]
	measureLabel := fieldDisplayLabel(measure)

	values := map[string]map[string]float64{}
	rowSet := map[string]struct{}{}
	colSet := map[string]struct{}{}

	for _, row := range rows {
		rowKey := pivotKey(row, rowFields)
		colKey := pivotKey(row, colFields)
		if colKey == "" {
			colKey = "Total"
		}
		rowSet[rowKey] = struct{}{}
		colSet[colKey] = struct{}{}
		if values[rowKey] == nil {
			values[rowKey] = map[string]float64{}
		}
		values[rowKey][colKey] += pivotMeasure(row, measure.Name)
	}

	out := &PivotMeta{
		RowLabels:    pivotSortedKeys(rowSet),
		ColLabels:    pivotSortedKeys(colSet),
		Values:       values,
		MeasureLabel: measureLabel,
	}
	if 0 == len(out.RowLabels) {
		out.RowLabels = []string{"Total"}
	}
	return out
}

func fieldDisplayLabel(f parser.Field) string {
	if s := strings.TrimSpace(f.Label); s != "" {
		return s
	}
	return strings.TrimSpace(f.Name)
}

func pivotKey(row map[string]interface{}, fields []parser.Field) string {
	if 0 == len(fields) {
		return "Total"
	}
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		parts = append(parts, pivotCellLabel(row, f.Name))
	}
	return strings.Join(parts, " / ")
}

func pivotCellLabel(row map[string]interface{}, fieldName string) string {
	if v, ok := row[fieldName+"_name"]; ok && orm.AsString(v) != "" {
		return orm.AsString(v)
	}
	v := row[fieldName]
	if v == nil {
		return "—"
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func pivotMeasure(row map[string]interface{}, fieldName string) float64 {
	v := row[fieldName]
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f
	default:
		f, _ := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(v)), 64)
		return f
	}
}

func pivotSortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

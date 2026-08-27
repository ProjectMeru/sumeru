package swcmeta

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// KanbanGroupExpander loads column definitions for grouped kanban boards.
type KanbanGroupExpander func(ctx context.Context, model, groupField string, records []map[string]interface{}) ([]KanbanColumn, error)

var kanbanGroupExpanders = map[string]KanbanGroupExpander{}

// RegisterKanbanGroupExpander registers a custom column loader for model:groupField.
func RegisterKanbanGroupExpander(model, groupField string, fn KanbanGroupExpander) {
	key := kanbanExpanderKey(model, groupField)
	kanbanGroupExpanders[key] = fn
}

func kanbanExpanderKey(model, groupField string) string {
	return strings.TrimSpace(model) + ":" + strings.TrimSpace(groupField)
}

// BuildKanbanColumns groups records into columns for a kanban view.
// groupOverride, when set, replaces the view default_group_by field.
func BuildKanbanColumns(ctx context.Context, view *parser.View, rows []map[string]interface{}, groupOverride string) ([]KanbanColumn, string, bool) {
	if view == nil {
		return nil, "", false
	}
	groupField := strings.TrimSpace(groupOverride)
	if groupField == "" {
		groupField = view.KanbanGroupField()
	}
	if groupField == "" {
		return nil, "", false
	}
	if rows == nil {
		rows = []map[string]interface{}{}
	}

	key := kanbanExpanderKey(view.Model, groupField)
	if fn, ok := kanbanGroupExpanders[key]; ok && fn != nil {
		cols, err := fn(ctx, view.Model, groupField, rows)
		if err == nil && len(cols) > 0 {
			return cols, groupField, view.KanbanDraggable()
		}
	}

	cols, err := defaultMany2OneKanbanColumns(ctx, view.Model, groupField, rows)
	if err != nil || len(cols) == 0 {
		return nil, groupField, view.KanbanDraggable()
	}
	return cols, groupField, view.KanbanDraggable()
}

func defaultMany2OneKanbanColumns(ctx context.Context, model, groupField string, rows []map[string]interface{}) ([]KanbanColumn, error) {
	comodel, ok := many2OneComodel(model, groupField)
	if !ok {
		return nil, fmt.Errorf("field %s is not many2one on %s", groupField, model)
	}

	comodelRows, err := orm.Search(ctx, comodel, [][]interface{}{})
	if err != nil {
		return nil, err
	}
	sortComodelRows(comodelRows)

	cols := make([]KanbanColumn, 0, len(comodelRows)+1)
	index := map[int64]int{}
	for _, cr := range comodelRows {
		id, ok := orm.CoerceInt64(cr["id"])
		if !ok || id <= 0 {
			continue
		}
		fold := asBoolRow(cr["fold"])
		if fold && !columnHasRecords(rows, groupField, id) {
			continue
		}
		cols = append(cols, KanbanColumn{
			Value:    id,
			Label:    columnLabel(cr),
			Sequence: intFromRow(cr["sequence"]),
			Color:    intFromRow(cr["color"]),
			Fold:     fold,
		})
		index[id] = len(cols) - 1
	}

	for _, row := range rows {
		gv, _ := orm.CoerceInt64(row[groupField])
		if gv <= 0 {
			continue
		}
		if ix, ok := index[gv]; ok {
			cols[ix].Records = append(cols[ix].Records, row)
		} else {
			cols = append(cols, KanbanColumn{
				Value:   gv,
				Label:   fmt.Sprintf("#%d", gv),
				Records: []map[string]interface{}{row},
			})
			index[gv] = len(cols) - 1
		}
	}

	var unassigned []map[string]interface{}
	for _, row := range rows {
		gv, _ := orm.CoerceInt64(row[groupField])
		if gv <= 0 {
			unassigned = append(unassigned, row)
		}
	}
	if len(unassigned) > 0 {
		cols = append([]KanbanColumn{{
			Value:    0,
			Label:    "Unassigned",
			Sequence: -1,
			Records:  unassigned,
		}}, cols...)
	}

	return cols, nil
}

func many2OneComodel(model, fieldName string) (string, bool) {
	inst, ok := orm.Registry[model]
	if !ok {
		return "", false
	}
	for _, f := range inst.Fields() {
		if f.Name == fieldName && f.Type == orm.Many2One {
			return strings.TrimSpace(f.Relation), f.Relation != ""
		}
	}
	return "", false
}

func sortComodelRows(rows []map[string]interface{}) {
	sort.SliceStable(rows, func(i, j int) bool {
		si := intFromRow(rows[i]["sequence"])
		sj := intFromRow(rows[j]["sequence"])
		if si != sj {
			return si < sj
		}
		return columnLabel(rows[i]) < columnLabel(rows[j])
	})
}

func columnLabel(row map[string]interface{}) string {
	if n := strings.TrimSpace(orm.AsString(row["name"])); n != "" {
		return n
	}
	id, _ := orm.CoerceInt64(row["id"])
	return fmt.Sprintf("#%d", id)
}

func columnHasRecords(rows []map[string]interface{}, groupField string, value int64) bool {
	for _, row := range rows {
		gv, _ := orm.CoerceInt64(row[groupField])
		if gv == value {
			return true
		}
	}
	return false
}

func intFromRow(v interface{}) int {
	n, ok := orm.CoerceInt64(v)
	if !ok {
		return 0
	}
	return int(n)
}

func asBoolRow(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case int:
		return t != 0
	case string:
		return t == "true" || t == "t" || t == "1"
	default:
		return false
	}
}

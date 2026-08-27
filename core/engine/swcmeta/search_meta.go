package swcmeta

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func isFilterableFieldType(t orm.FieldType) bool {
	switch t {
	case orm.Char, orm.Text, orm.Integer, orm.Float, orm.Float64, orm.Numeric,
		orm.Boolean, orm.Date, orm.DateTime, orm.Selection, orm.Many2One:
		return true
	default:
		return false
	}
}

func isGroupByFieldType(t orm.FieldType) bool {
	switch t {
	case orm.Char, orm.Boolean, orm.Date, orm.DateTime, orm.Selection, orm.Many2One:
		return true
	default:
		return false
	}
}

func fieldDefToArch(fd orm.FieldDefinition) ArchField {
	f := ArchField{
		Name:   fd.Name,
		String: fd.String,
		Type:   string(fd.Type),
	}
	if fd.String == "" {
		f.String = fd.Name
	}
	if fd.Relation != "" {
		f.Relation = fd.Relation
	}
	if len(fd.Selection) > 0 {
		f.Selection = fd.Selection
	}
	return f
}

// FilterableFields returns ORM fields suitable for custom filter UI.
func FilterableFields(model string) []ArchField {
	model = strings.TrimSpace(model)
	inst, ok := orm.Registry[model]
	if !ok {
		return nil
	}
	var out []ArchField
	for _, fd := range inst.Fields() {
		if fd.Virtual || fd.Name == "id" {
			continue
		}
		if !isFilterableFieldType(fd.Type) {
			continue
		}
		out = append(out, fieldDefToArch(fd))
	}
	return out
}

// GroupByFields returns ORM fields suitable for group-by UI.
func GroupByFields(model string) []ArchField {
	model = strings.TrimSpace(model)
	inst, ok := orm.Registry[model]
	if !ok {
		return nil
	}
	var out []ArchField
	for _, fd := range inst.Fields() {
		if fd.Virtual || fd.Name == "id" {
			continue
		}
		if !isGroupByFieldType(fd.Type) {
			continue
		}
		out = append(out, fieldDefToArch(fd))
	}
	return out
}

func searchFieldsFromView(ctx context.Context, model string, view *parser.View) []ArchField {
	if view == nil || len(view.Field) == 0 {
		return nil
	}
	fields := serializeFields(ctx, view.Field)
	return enrichFields(model, fields)
}

// BuildSearchMeta merges search-view presets with model field catalogs.
func BuildSearchMeta(ctx context.Context, model string, searchView *parser.View) *SearchMeta {
	model = strings.TrimSpace(model)
	meta := &SearchMeta{
		FilterFields:  FilterableFields(model),
		GroupByFields: GroupByFields(model),
	}
	if searchView != nil {
		if presets := serializeSearchFilters(searchView); len(presets) > 0 {
			meta.Filters = presets
		}
		if sf := searchFieldsFromView(ctx, model, searchView); len(sf) > 0 {
			meta.SearchFields = sf
		}
	}
	if len(meta.Filters) == 0 && len(meta.FilterFields) == 0 && len(meta.GroupByFields) == 0 && len(meta.SearchFields) == 0 {
		return nil
	}
	return meta
}

func serializeSearchFilters(view *parser.View) []SearchFilterMeta {
	if view == nil || len(view.SearchFilter) == 0 {
		return nil
	}
	out := make([]SearchFilterMeta, 0, len(view.SearchFilter))
	for _, f := range view.SearchFilter {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		label := strings.TrimSpace(f.String)
		if label == "" {
			label = name
		}
		out = append(out, SearchFilterMeta{
			Name:    name,
			String:  label,
			Domain:  strings.TrimSpace(f.Domain),
			GroupBy: strings.TrimSpace(f.GroupBy),
		})
	}
	return out
}

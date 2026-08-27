package swcmeta

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
	"sumeru/core/report"
)

// SerializeView converts a parsed view arch into SWC JSON metadata (no user ACL filter).
func SerializeView(view *parser.View) ViewArch {
	return SerializeViewForUser(context.Background(), view)
}

// SerializeViewForUser converts a parsed view arch, dropping fields the user cannot see via groups=.
func SerializeViewForUser(ctx context.Context, view *parser.View) ViewArch {
	if view == nil {
		return ViewArch{}
	}
	model := strings.TrimSpace(view.Model)
	arch := ViewArch{
		Type:   strings.TrimSpace(view.Type),
		Model:  model,
		Title:  strings.TrimSpace(view.Title),
		Fields: enrichFields(model, serializeFields(ctx, view.Field)),
	}
	if view.Header != nil {
		arch.Header = &ArchHeader{
			Buttons: serializeButtons(view.Header.Button),
			Fields:  enrichFields(model, serializeFields(ctx, view.Header.Field)),
		}
	}
	if view.Footer != nil && len(view.Footer.Button) > 0 {
		arch.Footer = &ArchFooter{
			Buttons: serializeButtons(view.Footer.Button),
		}
	}
	if view.Sheet != nil {
		arch.Sheet = serializeSheet(ctx, model, view.Sheet)
	}
	arch.FormMeta = formMetaForModel(model)
	if view.Chatter != nil {
		arch.HasChatter = true
	}
	if strings.EqualFold(view.Type, "kanban") {
		arch.Kanban = &KanbanMeta{
			GroupField:  view.KanbanGroupField(),
			Draggable:   view.KanbanDraggable(),
			QuickCreate: view.KanbanQuickCreate(),
		}
	}
	if strings.EqualFold(view.Type, "graph") {
		arch.Graph = &GraphMeta{Chart: view.GraphChart()}
	}
	if strings.EqualFold(view.Type, "calendar") || strings.TrimSpace(view.DateStart) != "" {
		arch.Calendar = &CalendarMeta{
			DateStart: strings.TrimSpace(view.DateStart),
			DateStop:  strings.TrimSpace(view.DateStop),
		}
	}
	if len(view.SearchFilter) > 0 {
		arch.Search = serializeSearch(view)
	}
	if caps := report.CapabilitiesFromView(view); caps.HasDownload() || caps.BulkUpload {
		arch.Report = &ReportMeta{
			Download:  caps.HasDownload(),
			Upload:    caps.BulkUpload,
			PDFSizes:  strings.Join(caps.PDFSizes, ","),
			BulkModes: strings.Join(caps.BulkModes, ","),
		}
	}
	return arch
}

func serializeSearch(view *parser.View) *SearchMeta {
	if view == nil || len(view.SearchFilter) == 0 {
		return nil
	}
	out := &SearchMeta{Filters: make([]SearchFilterMeta, 0, len(view.SearchFilter))}
	for _, f := range view.SearchFilter {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		label := strings.TrimSpace(f.String)
		if label == "" {
			label = name
		}
		out.Filters = append(out.Filters, SearchFilterMeta{
			Name:    name,
			String:  label,
			Domain:  strings.TrimSpace(f.Domain),
			GroupBy: strings.TrimSpace(f.GroupBy),
		})
	}
	if len(out.Filters) == 0 {
		return nil
	}
	return out
}

func formMetaForModel(model string) *FormMeta {
	inst, ok := orm.Registry[model]
	if !ok {
		return &FormMeta{}
	}
	meta := &FormMeta{}
	for _, f := range inst.Fields() {
		if f.Name == "image" {
			meta.HasImageField = true
			break
		}
	}
	return meta
}

func enrichFields(model string, fields []ArchField) []ArchField {
	if model == "" || len(fields) == 0 {
		return fields
	}
	out := make([]ArchField, len(fields))
	for i, f := range fields {
		out[i] = enrichField(model, f)
	}
	return out
}

func enrichField(model string, f ArchField) ArchField {
	inst, ok := orm.Registry[model]
	if !ok {
		return f
	}
	var fd orm.FieldDefinition
	found := false
	for _, fieldDef := range inst.Fields() {
		if fieldDef.Name != f.Name {
			continue
		}
		fd = fieldDef
		found = true
		break
	}
	if !found {
		return f
	}
	if f.Type == "" {
		f.Type = string(fd.Type)
	}
	if f.String == "" {
		f.String = fd.String
	}
	if f.Relation == "" {
		f.Relation = fd.Relation
	}
	if len(f.Selection) == 0 && len(fd.Selection) > 0 {
		f.Selection = fd.Selection
	}
	if fd.Required {
		f.Required = true
	}
	if f.Widget == "" && fd.Widget != "" {
		f.Widget = fd.Widget
	}
	if f.Options == nil && fd.Domain != "" {
		f.Options = map[string]string{"domain": fd.Domain}
	}
	if f.Options == nil && fd.Relation != "" {
		f.Options = map[string]string{"relation": fd.Relation}
	}
	if fd.Type == orm.One2Many {
		f = enrichOne2ManyField(model, f, fd)
	}
	return f
}

func enrichOne2ManyField(parentModel string, f ArchField, fd orm.FieldDefinition) ArchField {
	comodel := strings.TrimSpace(fd.Relation)
	if comodel == "" {
		comodel = strings.TrimSpace(f.Relation)
	}
	if f.Options == nil {
		f.Options = map[string]string{}
	}
	if inv := orm.ResolveInverseOne2ManyField(parentModel, comodel); inv != "" {
		f.Options["inverse"] = inv
		f.Options["relation"] = comodel
	}
	if f.Subview != nil && len(f.Subview.Fields) > 0 {
		f.Subview.Fields = enrichFields(comodel, f.Subview.Fields)
		return f
	}
	f.Subview = &ArchListSubview{
		Editable: "bottom",
		Fields:   autoColumnsForComodel(parentModel, comodel),
	}
	return f
}

func autoColumnsForComodel(parentModel, comodel string) []ArchField {
	inst, ok := orm.Registry[comodel]
	if !ok || inst == nil {
		return nil
	}
	inv := orm.ResolveInverseOne2ManyField(parentModel, comodel)
	skip := map[string]bool{
		"id": true, "create_uid": true, "write_uid": true,
		"create_date": true, "write_date": true,
	}
	if inv != "" {
		skip[inv] = true
	}
	out := make([]ArchField, 0, 8)
	for _, fd := range inst.Fields() {
		if skip[fd.Name] {
			continue
		}
		switch fd.Type {
		case orm.Char, orm.Text, orm.Integer, orm.Float, orm.Numeric,
			orm.Selection, orm.Date, orm.DateTime, orm.Boolean, orm.Many2One:
			out = append(out, enrichField(comodel, ArchField{
				Name:   fd.Name,
				String: fd.String,
				Type:   string(fd.Type),
			}))
		}
		if len(out) >= 6 {
			break
		}
	}
	return out
}

func serializeSheet(ctx context.Context, model string, s *parser.Sheet) *ArchSheet {
	out := &ArchSheet{
		Fields:     enrichFields(model, serializeFields(ctx, s.Field)),
		Groups:     []ArchGroup{},
		Divs:       serializeDivs(ctx, model, s),
		Separators: serializeSeparators(s.Separator),
		Labels:     serializeLabels(s.Label),
	}
	for _, g := range s.Group {
		out.Groups = append(out.Groups, serializeGroup(ctx, model, g))
	}
	for _, nb := range s.Notebook {
		pages := make([]ArchPage, 0, len(nb.Page))
		for _, p := range nb.Page {
			pg := ArchPage{
				Title:      strings.TrimSpace(p.Title),
				Fields:     enrichFields(model, serializeFields(ctx, p.Field)),
				Groups:     []ArchGroup{},
				Separators: serializeSeparators(p.Separator),
				Labels:     serializeLabels(p.Label),
			}
			for _, g := range p.Group {
				pg.Groups = append(pg.Groups, serializeGroup(ctx, model, g))
			}
			pages = append(pages, pg)
		}
		out.Notebook = append(out.Notebook, ArchNotebook{Pages: pages})
	}
	return out
}

func serializeGroup(ctx context.Context, model string, g parser.Group) ArchGroup {
	out := ArchGroup{
		String:     strings.TrimSpace(g.Title),
		Col:        parseArchInt(g.Col),
		Colspan:    parseArchInt(g.Colspan),
		Fields:     enrichFields(model, serializeFields(ctx, g.Field)),
		Groups:     []ArchGroup{},
		Separators: serializeSeparators(g.Separator),
		Labels:     serializeLabels(g.Label),
	}
	for _, nested := range g.Group {
		out.Groups = append(out.Groups, serializeGroup(ctx, model, nested))
	}
	return out
}

func parseArchInt(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func serializeDivs(ctx context.Context, model string, s *parser.Sheet) []ArchDiv {
	if s == nil || len(s.Div) == 0 {
		return nil
	}
	out := make([]ArchDiv, 0, len(s.Div))
	for _, d := range s.Div {
		div := serializeDiv(ctx, model, d)
		if divHasContent(div) {
			out = append(out, div)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func serializeDiv(ctx context.Context, model string, d parser.Div) ArchDiv {
	out := ArchDiv{
		Class:   strings.TrimSpace(d.Class),
		Fields:  enrichFields(model, serializeFields(ctx, d.Field)),
		Buttons: serializeButtons(d.Button),
	}
	for _, h1 := range d.H1 {
		out.H1Fields = append(out.H1Fields, enrichFields(model, serializeFields(ctx, h1.Field))...)
	}
	for _, nested := range d.Div {
		child := serializeDiv(ctx, model, nested)
		if divHasContent(child) {
			out.Divs = append(out.Divs, child)
		}
	}
	return out
}

func divHasContent(d ArchDiv) bool {
	return len(d.Fields) > 0 || len(d.H1Fields) > 0 || len(d.Divs) > 0 || len(d.Buttons) > 0
}

func serializeButtons(buttons []parser.Button) []ArchButton {
	out := make([]ArchButton, 0, len(buttons))
	for _, b := range buttons {
		out = append(out, ArchButton{
			Name:   strings.TrimSpace(b.Name),
			String: strings.TrimSpace(b.String),
			Type:   strings.TrimSpace(b.Type),
			Class:  strings.TrimSpace(b.Class),
		})
	}
	return out
}

func serializeFields(ctx context.Context, fields []parser.Field) []ArchField {
	out := make([]ArchField, 0, len(fields))
	uid := orm.SecurityUID(ctx)
	for _, f := range fields {
		if !orm.UserHasAnyAccessGroup(ctx, uid, f.Groups) {
			continue
		}
		af := ArchField{
			Name:        strings.TrimSpace(f.Name),
			String:      strings.TrimSpace(f.Label),
			Widget:      strings.TrimSpace(f.Widget),
			Placeholder: strings.TrimSpace(f.Placeholder),
			PivotType:   strings.TrimSpace(f.PivotType),
			Options:     parseFieldOptions(f.Options),
		}
		if lit, truthy, expr := parser.AttrLiteralOrExpr(f.Readonly); lit {
			af.Readonly = truthy
		} else {
			af.ReadonlyExpr = expr
		}
		if lit, truthy, expr := parser.AttrLiteralOrExpr(f.Required); lit {
			af.Required = truthy
		} else {
			af.RequiredExpr = expr
		}
		if lit, truthy, expr := parser.AttrLiteralOrExpr(f.Invisible); lit {
			af.Invisible = truthy
		} else {
			af.InvisibleExpr = expr
		}
		sub := f.List
		if sub == nil {
			sub = f.Tree
		}
		if sub != nil {
			af.Subview = serializeFieldList(ctx, sub)
		}
		out = append(out, af)
	}
	return out
}

func serializeFieldList(ctx context.Context, list *parser.FieldList) *ArchListSubview {
	if list == nil {
		return nil
	}
	return &ArchListSubview{
		Editable: strings.TrimSpace(list.Editable),
		Fields:   serializeFields(ctx, list.Field),
	}
}

func parseFieldOptions(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if kv := strings.SplitN(part, ":", 2); len(kv) == 2 {
			out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func serializeSeparators(items []parser.Separator) []ArchSeparator {
	if len(items) == 0 {
		return nil
	}
	out := make([]ArchSeparator, 0, len(items))
	for _, s := range items {
		out = append(out, ArchSeparator{String: strings.TrimSpace(s.String)})
	}
	return out
}

func serializeLabels(items []parser.Label) []ArchLabel {
	if len(items) == 0 {
		return nil
	}
	out := make([]ArchLabel, 0, len(items))
	for _, l := range items {
		out = append(out, ArchLabel{
			For:    strings.TrimSpace(l.For),
			String: strings.TrimSpace(l.String),
		})
	}
	return out
}

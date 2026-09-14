package parser

import (
	"fmt"
	"strings"
)

// ValidateViewArch checks list/kanban/report bool attrs and field/button modifiers.
func ValidateViewArch(v *View) error {
	if v == nil {
		return nil
	}
	if err := applyListOpenFlag(v); err != nil {
		return err
	}
	if err := applyViewKanbanBools(v); err != nil {
		return err
	}
	if err := applyViewBulkUpload(v); err != nil {
		return err
	}
	ctx := viewArchContext{ID: strings.TrimSpace(v.ID), Model: strings.TrimSpace(v.Model)}
	if err := validateFields(v.Field, ctx); err != nil {
		return err
	}
	for _, g := range v.Group {
		if err := validateGroup(g, ctx); err != nil {
			return err
		}
	}
	if v.Header != nil {
		if err := validateButtons(v.Header.Button, ctx); err != nil {
			return err
		}
		if err := validateFields(v.Header.Field, ctx); err != nil {
			return err
		}
	}
	if v.Footer != nil {
		if err := validateButtons(v.Footer.Button, ctx); err != nil {
			return err
		}
	}
	if v.Sheet != nil {
		if err := validateSheet(v.Sheet, ctx); err != nil {
			return err
		}
	}
	if v.Chatter != nil {
		if err := validateFields(v.Chatter.Field, ctx); err != nil {
			return err
		}
	}
	return nil
}

type viewArchContext struct {
	ID    string
	Model string
}

func (c viewArchContext) wrap(err error) error {
	if err == nil {
		return nil
	}
	if c.ID != "" || c.Model != "" {
		return fmt.Errorf("%w (view id=%q model=%q)", err, c.ID, c.Model)
	}
	return err
}

func applyListOpenFlag(v *View) error {
	if v == nil || !strings.EqualFold(strings.TrimSpace(v.Type), "list") {
		return nil
	}
	open := strings.TrimSpace(v.ListOpenAttr)
	switch {
	case open == "", strings.EqualFold(open, "true"):
		v.ListNoRowOpen = false
	case strings.EqualFold(open, "false"):
		v.ListNoRowOpen = true
	default:
		err := fmt.Errorf("invalid open=%q (only \"true\" or \"false\" allowed)", open)
		if id, model := strings.TrimSpace(v.ID), strings.TrimSpace(v.Model); id != "" || model != "" {
			return fmt.Errorf("%w (view id=%q model=%q)", err, id, model)
		}
		return err
	}
	return nil
}

func applyViewKanbanBools(v *View) error {
	if strings.TrimSpace(v.QuickCreate) != "" {
		b, err := ParseXMLBoolAttr("quick_create", v.QuickCreate)
		if err != nil {
			return err
		}
		v.kanbanQuickCreate = b
	}
	if strings.TrimSpace(v.RecordsDraggable) != "" {
		b, err := ParseXMLBoolAttr("records_draggable", v.RecordsDraggable)
		if err != nil {
			return err
		}
		v.kanbanDraggable = b
	}
	return nil
}

func applyViewBulkUpload(v *View) error {
	if strings.TrimSpace(v.BulkUpload) == "" {
		v.bulkUpload = false
		return nil
	}
	b, err := ParseXMLBoolAttr("bulk_upload", v.BulkUpload)
	if err != nil {
		return err
	}
	v.bulkUpload = b
	return nil
}

func validateSheet(s *Sheet, ctx viewArchContext) error {
	if err := validateFields(s.Field, ctx); err != nil {
		return err
	}
	for _, g := range s.Group {
		if err := validateGroup(g, ctx); err != nil {
			return err
		}
	}
	for _, d := range s.Div {
		if err := validateDiv(d, ctx); err != nil {
			return err
		}
	}
	for _, nb := range s.Notebook {
		for _, p := range nb.Page {
			if err := validateFields(p.Field, ctx); err != nil {
				return err
			}
			for _, g := range p.Group {
				if err := validateGroup(g, ctx); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateGroup(g Group, ctx viewArchContext) error {
	if err := validateFields(g.Field, ctx); err != nil {
		return err
	}
	for _, nested := range g.Group {
		if err := validateGroup(nested, ctx); err != nil {
			return err
		}
	}
	return nil
}

func validateDiv(d Div, ctx viewArchContext) error {
	if err := validateFields(d.Field, ctx); err != nil {
		return err
	}
	if err := validateButtons(d.Button, ctx); err != nil {
		return err
	}
	for _, h1 := range d.H1 {
		if err := validateFields(h1.Field, ctx); err != nil {
			return err
		}
	}
	for _, nested := range d.Div {
		if err := validateDiv(nested, ctx); err != nil {
			return err
		}
	}
	return nil
}

func validateFields(fields []Field, ctx viewArchContext) error {
	for _, f := range fields {
		if err := validateField(f, ctx); err != nil {
			return err
		}
	}
	return nil
}

func validateField(f Field, ctx viewArchContext) error {
	for _, name := range []struct {
		attr string
		raw  string
	}{
		{"invisible", f.Invisible},
		{"readonly", f.Readonly},
		{"required", f.Required},
	} {
		if strings.TrimSpace(name.raw) == "" {
			continue
		}
		if _, _, _, err := ParseModifierAttr(name.attr, name.raw); err != nil {
			return ctx.wrap(err)
		}
	}
	sub := f.List
	if sub == nil {
		sub = f.Tree
	}
	if sub != nil {
		if err := validateFields(sub.Field, ctx); err != nil {
			return err
		}
	}
	return nil
}

func validateButtons(buttons []Button, ctx viewArchContext) error {
	for _, b := range buttons {
		if strings.TrimSpace(b.Invisible) == "" {
			continue
		}
		if _, _, _, err := ParseModifierAttr("invisible", b.Invisible); err != nil {
			return ctx.wrap(err)
		}
	}
	return nil
}

package swcmeta_test

import (
	"sumeru/core/engine/swcmeta"
	"context"
	"strings"
	"testing"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)



func TestSerializeSheetNestedGroupsAndDivs(t *testing.T) {
	sheet := &parser.Sheet{
		Div: []parser.Div{
			{
				Class: "sum_title",
				H1: []parser.H1{
					{Field: []parser.Field{{Name: "name", Label: "Name"}}},
				},
			},
		},
		Group: []parser.Group{
			{
				Group: []parser.Group{
					{
						Title: "Contact",
						Field: []parser.Field{{Name: "email", Label: "Email"}},
					},
					{
						Title: "Address",
						Field: []parser.Field{{Name: "street", Label: "Street"}},
					},
				},
			},
		},
		Notebook: []parser.Notebook{
			{
				Page: []parser.Page{
					{
						Title: "Notes",
						Field: []parser.Field{{Name: "comment", Label: "Notes"}},
					},
				},
			},
		},
	}

	got := swcmeta.SerializeSheetForTest(context.Background(), "core.partner", sheet)
	if got == nil {
		t.Fatal("expected sheet")
	}
	if len(got.Divs) != 1 || len(got.Divs[0].H1Fields) != 1 || got.Divs[0].H1Fields[0].Name != "name" {
		t.Fatalf("unexpected divs: %+v", got.Divs)
	}
	if len(got.Groups) != 1 || len(got.Groups[0].Groups) != 2 {
		t.Fatalf("expected nested groups: %+v", got.Groups)
	}
	if got.Groups[0].Groups[0].String != "Contact" || got.Groups[0].Groups[0].Fields[0].Name != "email" {
		t.Fatalf("unexpected contact group: %+v", got.Groups[0].Groups[0])
	}
	if len(got.Notebook) != 1 || len(got.Notebook[0].Pages) != 1 {
		t.Fatalf("unexpected notebook: %+v", got.Notebook)
	}
	if got.Notebook[0].Pages[0].Fields[0].Name != "comment" {
		t.Fatalf("unexpected page fields: %+v", got.Notebook[0].Pages[0].Fields)
	}
}

func TestSerializeGroupColAttributes(t *testing.T) {
	g := swcmeta.SerializeGroupForTest(context.Background(), "crm.lead", parser.Group{
		Col: "2",
		Group: []parser.Group{
			{Title: "Left", Colspan: "1", Field: []parser.Field{{Name: "a"}}},
			{Title: "Right", Colspan: "1", Field: []parser.Field{{Name: "b"}}},
		},
	})
	if g.Col != 2 {
		t.Fatalf("expected col 2, got %d", g.Col)
	}
	if len(g.Groups) != 2 || g.Groups[0].Colspan != 1 || g.Groups[1].Colspan != 1 {
		t.Fatalf("unexpected nested groups: %+v", g.Groups)
	}
}

func TestSerializeGroupRecursive(t *testing.T) {
	g := swcmeta.SerializeGroupForTest(context.Background(), "core.partner", parser.Group{
		Title: "Outer",
		Group: []parser.Group{
			{Title: "Inner", Field: []parser.Field{{Name: "x"}}},
		},
	})
	if g.String != "Outer" || len(g.Groups) != 1 || g.Groups[0].Fields[0].Name != "x" {
		t.Fatalf("unexpected group: %+v", g)
	}
	if strings.TrimSpace(g.Groups[0].String) != "Inner" {
		t.Fatalf("expected inner title")
	}
}

func TestSerializeDivContactRow(t *testing.T) {
	div := swcmeta.SerializeDivForTest(context.Background(), "core.user", parser.Div{
		Class: "sum_title",
		H1: []parser.H1{
			{Field: []parser.Field{{Name: "name", Placeholder: "Name"}}},
		},
		Div: []parser.Div{
			{
				Class: "sum-title-contact-row",
				Field: []parser.Field{
					{Name: "email", Label: "Email", Widget: "email"},
					{Name: "phone", Label: "Phone"},
				},
			},
		},
	})
	if len(div.H1Fields) != 1 || div.H1Fields[0].Name != "name" {
		t.Fatalf("unexpected h1: %+v", div.H1Fields)
	}
	if len(div.Divs) != 1 || len(div.Divs[0].Fields) != 2 {
		t.Fatalf("unexpected contact row: %+v", div.Divs)
	}
}

func TestSerializeSheetSeparatorsAndLabels(t *testing.T) {
	sheet := &parser.Sheet{
		Separator: []parser.Separator{{String: "Section"}},
		Label:     []parser.Label{{For: "email", String: "Email hint"}},
	}
	got := swcmeta.SerializeSheetForTest(context.Background(), "my.module", sheet)
	if len(got.Separators) != 1 || got.Separators[0].String != "Section" {
		t.Fatalf("unexpected separators: %+v", got.Separators)
	}
	if len(got.Labels) != 1 || got.Labels[0].For != "email" {
		t.Fatalf("unexpected labels: %+v", got.Labels)
	}
}

func TestFormMetaHasImageField(t *testing.T) {
	const model = "test.formmeta.partner"
	orm.Registry[model] = stubImageModel{}
	t.Cleanup(func() { delete(orm.Registry, model) })

	meta := swcmeta.FormMetaForModelForTest(model)
	if meta == nil || !meta.HasImageField {
		t.Fatal("expected model to have image field")
	}
}

func TestSerializeFieldListSubview(t *testing.T) {
	fields := swcmeta.SerializeFieldsForTest(context.Background(), []parser.Field{
		{
			Name: "line_ids",
			List: &parser.FieldList{
				Editable: "bottom",
				Field: []parser.Field{
					{Name: "name", Label: "Description"},
					{Name: "quantity", Label: "Qty"},
				},
			},
		},
	})
	if len(fields) != 1 || fields[0].Subview == nil {
		t.Fatalf("expected subview: %+v", fields)
	}
	if fields[0].Subview.Editable != "bottom" || len(fields[0].Subview.Fields) != 2 {
		t.Fatalf("unexpected subview: %+v", fields[0].Subview)
	}
}

func TestSerializeModifierExpressions(t *testing.T) {
	fields := swcmeta.SerializeFieldsForTest(context.Background(), []parser.Field{
		{Name: "amount", Invisible: "state == 'done'", Readonly: "1"},
	})
	if len(fields) != 1 {
		t.Fatalf("fields: %+v", fields)
	}
	if fields[0].InvisibleExpr != "state == 'done'" || !fields[0].Readonly || fields[0].Invisible {
		t.Fatalf("modifiers: %+v", fields[0])
	}
}

func TestSerializeSearchFilters(t *testing.T) {
	view := &parser.View{
		Type: "search",
		SearchFilter: []parser.SearchFilter{
			{Name: "won", String: "Won", Domain: `[["won_status","=","won"]]`},
			{Name: "stage", String: "Stage", GroupBy: "stage_id"},
		},
	}
	out := swcmeta.SerializeView(view)
	if out.Search == nil || len(out.Search.Filters) != 2 {
		t.Fatalf("search: %+v", out.Search)
	}
}

func TestEnrichOne2ManyAutoColumns(t *testing.T) {
	const parent = "test.o2m.parent"
	const child = "test.o2m.child"
	orm.Registry[parent] = stubO2MParent{}
	orm.Registry[child] = stubO2MChild{}
	t.Cleanup(func() {
		delete(orm.Registry, parent)
		delete(orm.Registry, child)
	})

	got := swcmeta.EnrichFieldForTest(parent, swcmeta.ArchField{Name: "line_ids"})
	if got.Type != "one2many" {
		t.Fatalf("expected one2many, got %q", got.Type)
	}
	if got.Options["inverse"] != "parent_id" {
		t.Fatalf("expected inverse parent_id, got %q", got.Options["inverse"])
	}
	if got.Subview == nil || len(got.Subview.Fields) == 0 {
		t.Fatal("expected auto subview columns")
	}
	if got.Subview.Fields[0].Name != "name" {
		t.Fatalf("unexpected first column: %+v", got.Subview.Fields[0])
	}
}

type stubO2MParent struct{}

func (stubO2MParent) ModelName() string { return "test.o2m.parent" }

func (stubO2MParent) Fields() []orm.FieldDefinition {
	return []orm.FieldDefinition{
		{Name: "line_ids", Type: orm.One2Many, Relation: "test.o2m.child"},
	}
}

type stubO2MChild struct{}

func (stubO2MChild) ModelName() string { return "test.o2m.child" }

func (stubO2MChild) Fields() []orm.FieldDefinition {
	return []orm.FieldDefinition{
		{Name: "parent_id", Type: orm.Many2One, Relation: "test.o2m.parent"},
		{Name: "name", Type: orm.Char, String: "Description"},
		{Name: "quantity", Type: orm.Integer, String: "Qty"},
	}
}

type stubImageModel struct{}

func (stubImageModel) ModelName() string { return "test.formmeta.partner" }

func (stubImageModel) Fields() []orm.FieldDefinition {
	return []orm.FieldDefinition{{Name: "image", Type: orm.Text}}
}

package parser_test

import (
	"testing"

	"sumeru/core/engine/parser"
)

func TestParseKanbanDefaultGroupBy(t *testing.T) {
	arch := `<view model="crm.lead" type="kanban" default_group_by="stage_id" records_draggable="1">
		<field name="name"/>
	</view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.KanbanGroupField() != "stage_id" {
		t.Fatalf("group field: got %q", v.KanbanGroupField())
	}
	if !v.KanbanDraggable() {
		t.Fatal("expected draggable")
	}
}

func TestParseKanbanRootDefaultGroupBy(t *testing.T) {
	arch := `<kanban default_group_by="stage_id"><field name="name"/></kanban>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "kanban" || v.KanbanGroupField() != "stage_id" {
		t.Fatalf("got type=%q group=%q", v.Type, v.KanbanGroupField())
	}
}

func TestKanbanDraggableFalse(t *testing.T) {
	arch := `<view type="kanban" default_group_by="x" records_draggable="0"><field name="a"/></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.KanbanDraggable() {
		t.Fatal("expected not draggable")
	}
}

func TestKanbanColumnsPerRow(t *testing.T) {
	arch := `<view type="kanban" columns_per_row="6"><field name="name"/></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.KanbanColumnsPerRow(); got != 6 {
		t.Fatalf("columns per row: got %d want 6", got)
	}
	if got := (&parser.View{}).KanbanColumnsPerRow(); got != 4 {
		t.Fatalf("default columns per row: got %d want 4", got)
	}
}

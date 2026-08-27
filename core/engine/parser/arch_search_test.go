package parser_test

import (
	"testing"

	"sumeru/core/engine/parser"
)

func TestParseSearchArch(t *testing.T) {
	arch := `<search string="Leads"><field name="name"/><filter name="won" string="Won" domain='[["won_status","=","won"]]'/><filter name="by_stage" string="Stage" group_by="stage_id"/></search>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "search" {
		t.Fatalf("type: %s", v.Type)
	}
	if len(v.SearchFilter) != 2 || v.SearchFilter[0].Name != "won" || v.SearchFilter[1].GroupBy != "stage_id" {
		t.Fatalf("filters: %+v", v.SearchFilter)
	}
}

func TestParseGraphArch(t *testing.T) {
	arch := `<graph type="pie"><field name="stage_id" type="row"/><field name="expected_revenue" type="measure"/></graph>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "graph" || v.GraphChart() != "pie" {
		t.Fatalf("graph: type=%s chart=%s", v.Type, v.GraphChart())
	}
	if len(v.Field) != 2 || v.Field[1].PivotType != "measure" {
		t.Fatalf("fields: %+v", v.Field)
	}
}

func TestParseCalendarArch(t *testing.T) {
	arch := `<calendar date_start="date_deadline" date_stop="date_closed"><field name="name"/></calendar>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "calendar" || v.DateStart != "date_deadline" || v.DateStop != "date_closed" {
		t.Fatalf("calendar: %+v", v)
	}
}

func TestParsePivotRootArch(t *testing.T) {
	arch := `<pivot><field name="stage_id" type="row"/><field name="amount" type="measure"/></pivot>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "pivot" || len(v.Field) != 2 {
		t.Fatalf("pivot: %+v", v)
	}
}

func TestParseViewTypedSearch(t *testing.T) {
	arch := `<view type="search" model="crm.lead"><filter name="my" domain='[["user_id","=","$uid"]]'/></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "search" || len(v.SearchFilter) != 1 {
		t.Fatalf("view search: %+v", v)
	}
}

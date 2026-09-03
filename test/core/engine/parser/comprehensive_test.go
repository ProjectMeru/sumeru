package parser_test

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
)

func TestRecordFieldMap_precedence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		rec  parser.Record
		want map[string]string
	}{
		{
			name: "ref beats eval and body",
			rec: parser.Record{Field: []parser.RecordField{
				{Name: "x", Ref: "base.user_admin", Eval: "1", Body: "body"},
			}},
			want: map[string]string{"x": "base.user_admin"},
		},
		{
			name: "eval beats body",
			rec: parser.Record{Field: []parser.RecordField{
				{Name: "y", Eval: " True ", Body: "ignored"},
			}},
			want: map[string]string{"y": "True"},
		},
		{
			name: "body trimmed",
			rec: parser.Record{Field: []parser.RecordField{
				{Name: "z", Body: "  hello  "},
			}},
			want: map[string]string{"z": "hello"},
		},
		{
			name: "skip unnamed",
			rec: parser.Record{Field: []parser.RecordField{
				{Name: "", Body: "skip"},
				{Name: "a", Body: "ok"},
			}},
			want: map[string]string{"a": "ok"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parser.RecordFieldMap(tc.rec)
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("RecordFieldMap[%q] = %q want %q; full=%v", k, got[k], v, got)
				}
			}
		})
	}
}

func TestValidateModuleRoot_rejectsUnknown(t *testing.T) {
	t.Parallel()
	err := parser.ValidateModuleRoot(xml.Name{Local: "invalidroot"})
	if err == nil || !strings.Contains(err.Error(), "sumeru") {
		t.Fatalf("ValidateModuleRoot: %v", err)
	}
}

func TestPeekModuleXMLRootName_empty(t *testing.T) {
	t.Parallel()
	_, err := parser.PeekModuleXMLRootName([]byte("   "))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestActionToRecord_domainAndViewID(t *testing.T) {
	t.Parallel()
	a := parser.Action{
		ID: "act_x", Model: "sale.order", Name: "Orders", ViewMode: "list,form",
		Domain: `[("state","=","draft")]`, ViewID: "view_sale_list",
		Context: `{"default_type":"opportunity"}`,
	}
	rec := a.ToRecord()
	if rec.Model != "sys.action.window" || rec.ID != "act_x" {
		t.Fatalf("ToRecord meta: %+v", rec)
	}
	m := parser.RecordFieldMap(rec)
	if !strings.Contains(m["domain"], "draft") {
		t.Fatalf("domain: %q", m["domain"])
	}
	if !strings.Contains(m["context"], "view_id") || !strings.Contains(m["context"], "default_type") {
		t.Fatalf("context: %q", m["context"])
	}
}

func TestMergeMenuListData_noUpdate(t *testing.T) {
	t.Parallel()
	in := []byte(`<sumeru><data noupdate="1"><menuitem id="m1" name="Menu"/></data></sumeru>`)
	var ml parser.MenuList
	if err := xml.Unmarshal(in, &ml); err != nil {
		t.Fatal(err)
	}
	ml.MergeMenuListData()
	if !ml.NoUpdate || len(ml.MenuItems) != 1 {
		t.Fatalf("MergeMenuListData: noupdate=%v menus=%d", ml.NoUpdate, len(ml.MenuItems))
	}
}

func TestParseViewFromArch_allRoots(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		arch    string
		typeWant string
		check   func(t *testing.T, v *parser.View)
	}{
		{
			name: "graph", arch: `<graph chart="pie"><field name="amount"/></graph>`,
			typeWant: "graph",
			check: func(t *testing.T, v *parser.View) {
				if v.GraphChart() != "pie" {
					t.Fatalf("GraphChart: %q", v.GraphChart())
				}
			},
		},
		{
			name: "calendar", arch: `<calendar date_start="start" date_stop="stop"><field name="name"/></calendar>`,
			typeWant: "calendar",
			check: func(t *testing.T, v *parser.View) {
				if v.DateStart != "start" || v.DateStop != "stop" {
					t.Fatalf("dates: start=%q stop=%q", v.DateStart, v.DateStop)
				}
			},
		},
		{
			name: "gantt", arch: `<gantt date_start="d1" date_stop="d2"><field name="name"/></gantt>`,
			typeWant: "gantt",
		},
		{
			name: "map", arch: `<map latitude="lat" longitude="lng"><field name="name"/></map>`,
			typeWant: "map",
			check: func(t *testing.T, v *parser.View) {
				if v.Latitude != "lat" || v.Longitude != "lng" {
					t.Fatalf("coords: lat=%q lng=%q", v.Latitude, v.Longitude)
				}
			},
		},
		{
			name: "cohort", arch: `<cohort date_start="d" interval="week" measure="count"><field name="x"/></cohort>`,
			typeWant: "cohort",
			check: func(t *testing.T, v *parser.View) {
				if v.Interval != "week" || v.Measure != "count" {
					t.Fatalf("cohort attrs: interval=%q measure=%q", v.Interval, v.Measure)
				}
			},
		},
		{
			name: "pivot", arch: `<pivot string="Pivot"><field name="amount" type="measure"/></pivot>`,
			typeWant: "pivot",
		},
		{
			name: "search filters", arch: `<search string="Search"><filter name="open" domain="[('state','=','open')]"/><field name="name"/></search>`,
			typeWant: "search",
			check: func(t *testing.T, v *parser.View) {
				if len(v.SearchFilter) != 1 || v.SearchFilter[0].Name != "open" {
					t.Fatalf("filters: %+v", v.SearchFilter)
				}
			},
		},
		{
			name: "list open off", arch: `<list open="off"><field name="a"/></list>`,
			typeWant: "list",
			check: func(t *testing.T, v *parser.View) {
				if !v.ListNoRowOpen {
					t.Fatal("ListNoRowOpen expected true for open=off")
				}
			},
		},
		{
			name: "kanban quick create off", arch: `<kanban default_group_by="stage" quick_create="no"><field name="n"/></kanban>`,
			typeWant: "kanban",
			check: func(t *testing.T, v *parser.View) {
				if v.KanbanQuickCreate() {
					t.Fatal("KanbanQuickCreate should be false when quick_create=no")
				}
			},
		},
		{
			name: "kanban columns clamp", arch: `<kanban columns_per_row="99"><field name="n"/></kanban>`,
			typeWant: "kanban",
			check: func(t *testing.T, v *parser.View) {
				if got := v.KanbanColumnsPerRow(); got != 12 {
					t.Fatalf("KanbanColumnsPerRow clamp: got %d", got)
				}
			},
		},
		{
			name: "view report attrs", arch: `<view type="list" report_download="csv,pdf" bulk_upload="1"><field name="x"/></view>`,
			typeWant: "list",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := parser.ParseViewFromArch(tc.arch)
			if err != nil {
				t.Fatal(err)
			}
			if v.Type != tc.typeWant {
				t.Fatalf("type = %q want %q", v.Type, tc.typeWant)
			}
			if tc.check != nil {
				tc.check(t, v)
			}
		})
	}
}

func TestParseViewFromArch_errors(t *testing.T) {
	t.Parallel()
	if _, err := parser.ParseViewFromArch(""); err == nil {
		t.Fatal("empty arch should error")
	}
	if _, err := parser.ParseViewFromArch("<unknown/>"); err == nil {
		t.Fatal("unsupported root should error")
	}
}

func TestParseViewFromArch_nilViewMethods(t *testing.T) {
	t.Parallel()
	var v *parser.View
	if v.KanbanGroupField() != "" || v.KanbanDraggable() || v.KanbanQuickCreate() {
		t.Fatal("nil view kanban helpers should be safe")
	}
	if v.KanbanColumnsPerRow() != 4 || v.GraphChart() != "bar" {
		t.Fatal("nil view defaults")
	}
}

func TestParseViewList_invalidRoot(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.xml"
	if err := os.WriteFile(path, []byte(`<root></root>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := parser.ParseViewList(path); err == nil {
		t.Fatal("expected error for non-sumeru root")
	}
}

func TestIsFalsyAttr_table(t *testing.T) {
	t.Parallel()
	for _, v := range []string{"0", "false", "no", "off", " FALSE "} {
		if !parser.IsFalsyAttr(v) {
			t.Errorf("IsFalsyAttr(%q) = false", v)
		}
	}
	for _, v := range []string{"", "1", "yes"} {
		if parser.IsFalsyAttr(v) {
			t.Errorf("IsFalsyAttr(%q) = true", v)
		}
	}
}

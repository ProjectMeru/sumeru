package module_test

import (
	"strings"
	"testing"

	"sumeru/core/engine/viewinherit"
	"sumeru/core/module"
)

func TestPartnerFormInheritAddsCustomerRank(t *testing.T) {
	parent := `<view id="view_core_partner_form" model="core.partner" type="form"><sheet><group><group string="Contact"><field name="email" string="Email"></field><field name="phone" string="Phone"></field></group></group></sheet></view>`
	frag := `<xpath expr="//field[@name='phone']" position="after"><field name="customer_rank" string="Customer Rank"></field></xpath>`
	out, err := viewinherit.ApplyInheritArch(parent, frag)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `name="customer_rank"`) {
		t.Fatalf("merged arch missing customer_rank: %s", out)
	}
}

func TestInferSysViewTypeFromArch(t *testing.T) {
	tests := []struct {
		arch string
		want string
	}{
		{"<list string=\"X\"><field name=\"a\"/></list>", "list"},
		{"  <form><sheet/></form>", "form"},
		{"<kanban><field name=\"n\"/></kanban>", "kanban"},
		{"<search><filter name=\"a\" domain='[]'/></search>", "search"},
		{"<graph type=\"bar\"><field name=\"n\"/></graph>", "graph"},
		{"<calendar date_start=\"d\"><field name=\"n\"/></calendar>", "calendar"},
		{"<gantt date_start=\"d\"><field name=\"n\"/></gantt>", "gantt"},
		{"<map latitude=\"lat\" longitude=\"lng\"><field name=\"n\"/></map>", "map"},
		{"<cohort date_start=\"d\" interval=\"month\"><field name=\"n\"/></cohort>", "cohort"},
		{"<pivot><field name=\"n\" type=\"measure\"/></pivot>", "pivot"},
		{"<view type=\"list\" model=\"m\"><list/></view>", "list"},
		{"", ""},
		{"<unknown/>", ""},
	}
	for _, tt := range tests {
		if got := module.InferSysViewTypeFromArch(tt.arch); got != tt.want {
			t.Errorf("InferSysViewTypeFromArch(%q) = %q; want %q", tt.arch, got, tt.want)
		}
	}
}

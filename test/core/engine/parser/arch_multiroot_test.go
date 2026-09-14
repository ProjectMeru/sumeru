package parser_test

import (
	"os"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
)

func TestParseViewFromArch_formRoot(t *testing.T) {
	arch := `<form><header><button name="save" string="Save" type="object"/></header><sheet><group><field name="x" string="X"/></group></sheet></form>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "form" || v.Header == nil || v.Sheet == nil {
		t.Fatalf("unexpected view: %#v", v)
	}
}

func TestParseViewFromArch_nestedFormUnderView(t *testing.T) {
	arch := `<view id="v1" model="demo.model" type="form"><form><sheet><field name="x" string="X"/></sheet></form></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Sheet == nil || len(v.Sheet.Field) != 1 || v.Sheet.Field[0].Name != "x" {
		t.Fatalf("expected nested form sheet/field promoted, got %#v", v)
	}
	if v.Form != nil {
		t.Fatalf("expected Form cleared after promote, got %#v", v.Form)
	}
}

func TestParseViewList_nestedFormUnderView(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/form.xml"
	content := `<?xml version="1.0" encoding="utf-8"?>
<sumeru>
  <data>
    <view id="view_demo_form" model="demo.model" type="form">
      <form>
        <sheet>
          <field name="x" string="X"/>
        </sheet>
      </form>
    </view>
  </data>
</sumeru>
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	vl, err := parser.ParseViewList(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(vl.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(vl.Views))
	}
	v := vl.Views[0]
	if v.Sheet == nil || len(v.Sheet.Field) != 1 || v.Sheet.Field[0].Name != "x" {
		t.Fatalf("expected nested form promoted in ParseViewList, got %#v", v)
	}
	if v.Form != nil {
		t.Fatalf("expected Form cleared after promote, got %#v", v.Form)
	}
}

func TestParseViewFromArch_listRoot(t *testing.T) {
	arch := `<list><field name="a"/><field name="b" string="B"/></list>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "list" || len(v.Field) != 2 {
		t.Fatalf("unexpected view: %#v", v)
	}
}

func TestParseViewFromArch_viewRootListOpenFalse(t *testing.T) {
	arch := `<view type="list" open="false"><field name="x" string="X"/></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "list" || len(v.Field) != 1 {
		t.Fatalf("unexpected view: %#v", v)
	}
	if !v.ListNoRowOpen {
		t.Fatalf("expected ListNoRowOpen for view root with open=false")
	}
}

func TestParseViewFromArch_listOpenFalse(t *testing.T) {
	arch := `<list open="false"><field name="a"/></list>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if !v.ListNoRowOpen {
		t.Fatalf("expected ListNoRowOpen true for open=false, got %#v", v)
	}
}

func TestParseViewFromArch_listOpenDefault(t *testing.T) {
	arch := `<list><field name="a"/></list>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.ListNoRowOpen {
		t.Fatalf("expected ListNoRowOpen false by default, got %#v", v)
	}
}

func TestParseViewFromArch_listOpenInvalid(t *testing.T) {
	_, err := parser.ParseViewFromArch(`<list open="off"><field name="a"/></list>`)
	if err == nil {
		t.Fatal("expected error for open=off")
	}
	msg := err.Error()
	if !strings.Contains(msg, "off") || !strings.Contains(msg, "true") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseViewFromArch_invisibleLegacyRejected(t *testing.T) {
	_, err := parser.ParseViewFromArch(`<form><field name="a" invisible="1"/></form>`)
	if err == nil {
		t.Fatal("expected error for invisible=1")
	}
	if !strings.Contains(err.Error(), "1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseViewFromArch_listOpenInvalidWithViewMeta(t *testing.T) {
	arch := `<view type="list" id="sale.order.list" model="sale.order" open="0"><field name="a"/></view>`
	_, err := parser.ParseViewFromArch(arch)
	if err == nil {
		t.Fatal("expected error for open=0")
	}
	msg := err.Error()
	for _, want := range []string{"0", "sale.order.list", "sale.order"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q missing %q", msg, want)
		}
	}
}

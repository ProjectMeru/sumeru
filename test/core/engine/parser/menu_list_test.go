package parser_test

import (
	"os"
	"testing"

	"sumeru/core/engine/parser"
)

func TestParseMenuList_mergesData(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/menus.xml"
	content := `<?xml version="1.0"?>
<sumeru>
  <data noupdate="true">
    <menuitem id="menu_root" name="Root" sequence="1"/>
    <action id="action_test" type="window" model="core.partner" name="Partners" view_mode="list,form"/>
  </data>
  <menuitem id="menu_leaf" name="Leaf" parent="menu_root" action="action_test"/>
</sumeru>`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ml, err := parser.ParseMenuList(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ml.NoUpdate {
		t.Fatal("expected noupdate")
	}
	if len(ml.MenuItems) != 2 {
		t.Fatalf("menus=%d", len(ml.MenuItems))
	}
	if len(ml.Actions) != 1 || len(ml.Records) != 1 {
		t.Fatalf("actions=%d records=%d", len(ml.Actions), len(ml.Records))
	}
}

package parser_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
)

func TestModuleXML_dataWrapperSumeru(t *testing.T) {
	in := []byte(`<sumeru><data><record id="x" model="y"><field name="a">b</field></record></data></sumeru>`)
	root, err := parser.PeekModuleXMLRootName(in)
	if err != nil || root != "sumeru" {
		t.Fatalf("root %q err %v", root, err)
	}
	if err := parser.ValidateModuleRoot(xml.Name{Local: root}); err != nil {
		t.Fatal(err)
	}
	var vl parser.ViewList
	if err := xml.Unmarshal(in, &vl); err != nil {
		t.Fatal(err)
	}
	vl.MergeViewListData()
	if len(vl.Records) != 1 || vl.Records[0].ID != "x" {
		t.Fatalf("records: %+v", vl.Records)
	}
}

func TestPeekModuleXMLRootName_sumeru(t *testing.T) {
	root, err := parser.PeekModuleXMLRootName([]byte("\n  <sumeru>\n</sumeru>"))
	if err != nil || root != "sumeru" {
		t.Fatalf("got %q err %v", root, err)
	}
}

func TestActionHelpNestedInnerXML(t *testing.T) {
	in := []byte(`<sumeru><data>
		<action id="action_contacts" type="window" model="core.partner" name="Contacts" view_mode="list,form,kanban">
			<help>
				<p class="sum-view-nocontent-smiling-face">Create your first Contact!</p>
			</help>
		</action>
		<menuitem id="menu_contacts" name="Contacts" action="action_contacts" sequence="1"/>
	</data></sumeru>`)
	var vl parser.ViewList
	if err := xml.Unmarshal(in, &vl); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	vl.MergeViewListData()
	if len(vl.Actions) != 1 {
		t.Fatalf("actions=%d", len(vl.Actions))
	}
	rec := vl.Actions[0].ToRecord()
	if rec.Model != "sys.action.window" || rec.ID != "action_contacts" {
		t.Fatalf("record: %+v", rec)
	}
	help := ""
	for _, f := range rec.Field {
		if f.Name == "help" {
			help = f.Body
		}
	}
	if !strings.Contains(help, "Create your first Contact!") {
		t.Fatalf("help body missing: %q", help)
	}
	if len(vl.MenuItems) != 1 || vl.MenuItems[0].ID != "menu_contacts" {
		t.Fatalf("menus: %+v", vl.MenuItems)
	}
	if len(vl.Records) != 1 {
		t.Fatalf("merged action records=%d", len(vl.Records))
	}
}

func TestActionSearchViewIDInContext(t *testing.T) {
	in := []byte(`<sumeru><data>
		<action id="action_crm" type="window" model="crm.lead" name="Pipeline" view_mode="kanban,list,form"
			search_view_id="view_crm_pipeline_search" context='{"default_type":"opportunity"}'/>
	</data></sumeru>`)
	var vl parser.ViewList
	if err := xml.Unmarshal(in, &vl); err != nil {
		t.Fatal(err)
	}
	vl.MergeViewListData()
	if len(vl.Actions) != 1 {
		t.Fatalf("actions=%d", len(vl.Actions))
	}
	rec := vl.Actions[0].ToRecord()
	var ctx string
	for _, f := range rec.Field {
		if f.Name == "context" {
			ctx = f.Body
		}
	}
	if !strings.Contains(ctx, "search_view_id") || !strings.Contains(ctx, "default_type") {
		t.Fatalf("context missing search_view_id or default_type: %q", ctx)
	}
}

func TestActionURLTypeToRecord(t *testing.T) {
	in := []byte(`<sumeru><data>
		<action id="action_report_pl" type="url" name="Profit &amp; Loss"
			url="/account/reports/view?type=profit_loss"/>
	</data></sumeru>`)
	var vl parser.ViewList
	if err := xml.Unmarshal(in, &vl); err != nil {
		t.Fatal(err)
	}
	vl.MergeViewListData()
	if len(vl.Actions) != 1 {
		t.Fatalf("actions=%d", len(vl.Actions))
	}
	rec := vl.Actions[0].ToRecord()
	if rec.Model != "sys.action.url" || rec.ID != "action_report_pl" {
		t.Fatalf("record: %+v", rec)
	}
	var name, url string
	for _, f := range rec.Field {
		switch f.Name {
		case "name":
			name = f.Body
		case "url":
			url = f.Body
		}
	}
	if name != "Profit & Loss" {
		t.Fatalf("name=%q", name)
	}
	if url != "/account/reports/view?type=profit_loss" {
		t.Fatalf("url=%q", url)
	}
	if len(vl.Records) != 1 || vl.Records[0].Model != "sys.action.url" {
		t.Fatalf("merged records: %+v", vl.Records)
	}
}

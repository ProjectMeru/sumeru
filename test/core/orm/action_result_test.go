package orm_test

import (
	"sumeru/core/orm"
	"testing"
)



func TestEncodeActionResultEmpty(t *testing.T) {
	got := orm.EncodeActionResult("")
	if got != true {
		t.Fatalf("empty redirect: %#v", got)
	}
}

func TestEncodeActionResultClose(t *testing.T) {
	got, ok := orm.EncodeActionResult("close").(map[string]interface{})
	if !ok || got["close"] != true {
		t.Fatalf("close: %#v", got)
	}
}

func TestEncodeActionResultOpenDialog(t *testing.T) {
	got, ok := orm.EncodeActionResult("/web?model=crm.lead.lost&id=12&view_type=form").(map[string]interface{})
	if !ok {
		t.Fatalf("type: %#v", got)
	}
	open, _ := got["open"].(map[string]interface{})
	if open["model"] != "crm.lead.lost" || open["recordId"] != 12 || open["target"] != "dialog" {
		t.Fatalf("open: %#v", open)
	}
}

func TestEncodeActionResultRedirect(t *testing.T) {
	got, ok := orm.EncodeActionResult("/web?action=5&menu_id=2").(map[string]interface{})
	if !ok || got["redirect"] != "/web?action=5&menu_id=2" {
		t.Fatalf("redirect: %#v", got)
	}
}

func TestActionOpenURL(t *testing.T) {
	u := orm.ActionOpenURL("crm.lead.lost", 9)
	if u != "/web?id=9&model=crm.lead.lost&target=dialog&view_type=form" {
		t.Fatalf("url: %s", u)
	}
}

package orm_test

import (
	"sumeru/core/orm"
	"testing"
)

func TestMenuRecordActionID(t *testing.T) {
	actionID, ok := orm.MenuRecordActionID(map[string]interface{}{"action_id": int64(7)})
	if !ok || actionID != 7 {
		t.Fatalf("got actionID=%d ok=%v", actionID, ok)
	}
	if _, ok := orm.MenuRecordActionID(map[string]interface{}{"action_id": int64(0)}); ok {
		t.Fatal("zero action_id should be rejected")
	}
}

func TestParseMenuIDString(t *testing.T) {
	menuID, ok := orm.ParseMenuIDString(" 12 ")
	if !ok || menuID != 12 {
		t.Fatalf("got id=%d ok=%v", menuID, ok)
	}
	if _, ok := orm.ParseMenuIDString("0"); ok {
		t.Fatal("zero menu id should be rejected")
	}
}

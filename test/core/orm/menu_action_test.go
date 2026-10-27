package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
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

func TestMenuEntryActionForUser_zeroUIDReturnsZero(t *testing.T) {
	actionID, leafID := orm.MenuEntryActionForUser(context.Background(), 0, 1)
	if actionID != 0 || leafID != 0 {
		t.Fatalf("uid=0 should not resolve menu entry; got action=%d leaf=%d", actionID, leafID)
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

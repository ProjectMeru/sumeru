package web_test

import (
	"sumeru/core/server/web"
	"testing"
)

func TestParseMenuIDString(t *testing.T) {
	menuID, ok := web.ParseMenuIDString(" 12 ")
	if !ok || menuID != 12 {
		t.Fatalf("got id=%d ok=%v", menuID, ok)
	}
	if _, ok := web.ParseMenuIDString(""); ok {
		t.Fatal("empty menu id should be rejected")
	}
	if _, ok := web.ParseMenuIDString("0"); ok {
		t.Fatal("zero menu id should be rejected")
	}
	if _, ok := web.ParseMenuIDString("abc"); ok {
		t.Fatal("non-numeric menu id should be rejected")
	}
}

func TestResolveActionIDFromQuery(t *testing.T) {
	ctx := t.Context()

	if got := web.ResolveActionIDFromQuery(ctx, " 42 "); got != 42 {
		t.Fatalf("numeric action id = %d, want 42", got)
	}
	if got := web.ResolveActionIDFromQuery(ctx, ""); got != 0 {
		t.Fatalf("empty action query should return 0, got %d", got)
	}
	if got := web.ResolveActionIDFromQuery(ctx, "not-a-number-or-xml"); got != 0 {
		t.Fatalf("unknown action reference should return 0, got %d", got)
	}
}

func TestMenuRecordActionID(t *testing.T) {
	actionID, ok := web.MenuRecordActionID(map[string]interface{}{"action_id": int64(7)})
	if !ok || actionID != 7 {
		t.Fatalf("got actionID=%d ok=%v", actionID, ok)
	}
	if _, ok := web.MenuRecordActionID(map[string]interface{}{"action_id": int64(0)}); ok {
		t.Fatal("zero action_id should be rejected")
	}
	if _, ok := web.MenuRecordActionID(map[string]interface{}{}); ok {
		t.Fatal("missing action_id should be rejected")
	}
}

func TestResolveWindowActionID_prefersActionQuery(t *testing.T) {
	ctx := t.Context()
	if got := web.ResolveWindowActionID(ctx, "9", "1"); got != 9 {
		t.Fatalf("action query should win over menu query, got %d", got)
	}
}

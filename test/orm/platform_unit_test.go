package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func TestHashAPIKey_stable(t *testing.T) {
	a := orm.HashAPIKey("sk_test")
	b := orm.HashAPIKey("sk_test")
	if a == "" || a != b {
		t.Fatalf("hash unstable: %q %q", a, b)
	}
	if a == orm.HashAPIKey("sk_other") {
		t.Fatal("different keys should hash differently")
	}
}

func TestGenerateAPIKey_format(t *testing.T) {
	raw, prefix, hash, err := orm.GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 10 || raw[:3] != "sk_" {
		t.Fatalf("raw = %q", raw)
	}
	if prefix == "" || hash == "" || hash != orm.HashAPIKey(raw) {
		t.Fatalf("prefix/hash mismatch: %q %q", prefix, hash)
	}
}

func TestSubstituteDomainContext_ABAC(t *testing.T) {
	dom := [][]interface{}{
		{"user_id", "=", "$uid"},
		{"company_id", "=", "$company_id"},
		{"company_id", "in", "$company_ids"},
		{"name", "=", "keep"},
	}
	out := orm.SubstituteDomainContext(dom, orm.DomainContext{
		UID:        9,
		CompanyID:  3,
		CompanyIDs: []int64{3, 5},
	})
	if out[0][2].(int64) != 9 {
		t.Fatalf("uid: %v", out[0][2])
	}
	if out[1][2].(int64) != 3 {
		t.Fatalf("company_id: %v", out[1][2])
	}
	arr, ok := out[2][2].([]interface{})
	if !ok || len(arr) != 2 {
		t.Fatalf("company_ids: %v", out[2][2])
	}
	if out[3][2] != "keep" {
		t.Fatalf("literal changed: %v", out[3][2])
	}
}

func TestRecordMatchesDomain(t *testing.T) {
	rec := map[string]interface{}{"state": "draft", "user_id": int64(1)}
	if !orm.RecordMatchesDomain(rec, [][]interface{}{{"state", "=", "draft"}}) {
		t.Fatal("expected match")
	}
	if orm.RecordMatchesDomain(rec, [][]interface{}{{"state", "=", "done"}}) {
		t.Fatal("expected no match")
	}
	if !orm.RecordMatchesDomain(rec, [][]interface{}{{"user_id", "in", []interface{}{int64(1), int64(2)}}}) {
		t.Fatal("expected in match")
	}
}

func TestValidatePasswordPolicy_minLength(t *testing.T) {
	// Without DB, GetConfig returns default "8".
	if err := orm.ValidatePasswordPolicy("short"); err == nil {
		t.Fatal("expected short password error")
	}
	if err := orm.ValidatePasswordPolicy("longenough"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestTranslate_fallbackWithoutDB(t *testing.T) {
	got := orm.Translate(orm.BackgroundBypass(), "hi_IN", "Settings")
	if got != "Settings" {
		t.Fatalf("expected source fallback, got %q", got)
	}
	if orm.Translate(orm.BackgroundBypass(), "en_US", "X") != "X" {
		t.Fatal("en_US should return src")
	}
}

func TestCanWorkflowTransition_bypass(t *testing.T) {
	ctx := orm.ContextWithBypass(orm.BackgroundBypass(), true)
	if err := orm.CanWorkflowTransition(ctx, orm.WorkflowTransitionInput{
		Model: "x.model", RecordID: 1, FromState: "a", ToState: "b", UID: 2,
	}); err != nil {
		t.Fatalf("bypass should allow: %v", err)
	}
}

func TestUIDFromAPIKey_empty(t *testing.T) {
	if orm.UIDFromAPIKey(orm.BackgroundBypass(), "") != 0 {
		t.Fatal("empty key should resolve to 0")
	}
	if orm.UIDFromAPIKey(orm.BackgroundBypass(), "sk_missing") != 0 {
		t.Fatal("without DB should resolve to 0")
	}
}

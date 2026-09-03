package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestRedactRecordForRead_stripsPassword(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	rec := map[string]interface{}{
		"id":       1,
		"login":    "admin",
		"password": "hash",
	}
	orm.RedactRecordForRead(ctx, 1, "core.user", rec)
	if _, ok := rec["password"]; ok {
		t.Fatal("password should be redacted from core.user reads")
	}
	if rec["login"] != "admin" {
		t.Fatalf("login=%v", rec["login"])
	}
}

func TestRedactRecordForRead_stripsKeyHash(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	rec := map[string]interface{}{"id": 1, "key_hash": "secret", "name": "k1"}
	orm.RedactRecordForRead(ctx, 1, "core.user.apikey", rec)
	if _, ok := rec["key_hash"]; ok {
		t.Fatal("key_hash should be redacted")
	}
}

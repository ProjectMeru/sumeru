package orm_test

import (
	"strings"
	"testing"
	"sumeru/core/orm"
)


func TestCriteriaToDomain_sortedEquals(t *testing.T) {
	d := orm.CriteriaToDomain(map[string]interface{}{"id": 7, "name": "base"})
	if len(d) != 2 {
		t.Fatalf("len=%d", len(d))
	}
	if d[0][0] != "id" || d[0][1] != "=" || d[0][2] != 7 {
		t.Fatalf("first triple: %v", d[0])
	}
	if d[1][0] != "name" || d[1][1] != "=" || d[1][2] != "base" {
		t.Fatalf("second triple: %v", d[1])
	}
	if orm.CriteriaToDomain(nil) != nil {
		t.Fatal("empty criteria should be nil")
	}
}

func TestBuildWhereWithRecordRules_searchOneCriteria(t *testing.T) {
	ctx := orm.ContextWithBypass(orm.BackgroundBypass(), true)
	domain := orm.CriteriaToDomain(map[string]interface{}{"id": 7, "name": "base"})
	sql, args, err := orm.BuildWhereWithRecordRules(ctx, 0, "sys.module", "read", domain)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, `"id"`) || !strings.Contains(sql, `"name"`) {
		t.Fatalf("expected quoted id and name in %s", sql)
	}
	if strings.Contains(sql, "FOR UPDATE") {
		t.Fatalf("search where should not lock: %s", sql)
	}
	if len(args) != 2 || args[0] != 7 || args[1] != "base" {
		t.Fatalf("args=%v", args)
	}
}

func TestFindUIDefaultViewDomain_compiles(t *testing.T) {
	ctx := orm.ContextWithBypass(orm.BackgroundBypass(), true)
	domain := [][]interface{}{
		{"model", "=", "sys.module"},
		{"type", "=", "list"},
	}
	sql, args, err := orm.BuildWhereWithRecordRules(ctx, 0, "sys.view", "read", domain)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, `"model"`) || !strings.Contains(sql, `"type"`) {
		t.Fatalf("sql=%s", sql)
	}
	if len(args) != 2 {
		t.Fatalf("args=%v", args)
	}
}

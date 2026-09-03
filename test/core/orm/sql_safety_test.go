package orm_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/orm"
)

type safetyModel struct {
	name   string
	fields []orm.FieldDefinition
}

func (m safetyModel) ModelName() string         { return m.name }
func (m safetyModel) Fields() []orm.FieldDefinition { return m.fields }

func registerSafetyModel(t *testing.T) {
	t.Helper()
	m := safetyModel{
		name:   "test.safety",
		fields: []orm.FieldDefinition{{Name: "name", Type: orm.Char}},
	}
	orm.RegisterModelWithModule(m, "test")
	t.Cleanup(func() { delete(orm.Registry, "test.safety") })
}

func TestQuotedPermColumnForOp(t *testing.T) {
	for _, op := range []string{"read", "write", "create", "unlink", "READ"} {
		col, err := orm.QuotedPermColumnForOp(op)
		if err != nil {
			t.Fatalf("op %q: %v", op, err)
		}
		if !strings.HasPrefix(col, `"perm_`) || !strings.HasSuffix(col, `"`) {
			t.Fatalf("op %q: expected quoted perm column, got %q", op, col)
		}
	}
	bad := []string{"", "drop", "1=1", "read;drop", "perm_evil"}
	for _, op := range bad {
		if _, err := orm.QuotedPermColumnForOp(op); err == nil {
			t.Fatalf("expected error for op %q", op)
		}
	}
}

func TestQuotedTableName_rejectsInjection(t *testing.T) {
	bad := []string{
		"core.user; DROP TABLE core_user--",
		"sys.module' OR '1'='1",
		"mail.message); DELETE FROM mail_message; --",
		"1bad",
	}
	for _, name := range bad {
		if _, err := orm.QuotedTableName(name); err == nil {
			t.Fatalf("expected reject for model %q", name)
		}
	}
}

func TestBuildWhereWithRecordRules_rejectsBadDomain(t *testing.T) {
	registerSafetyModel(t)
	ctx := orm.ContextWithBypass(context.Background(), true)

	badDomains := [][][]interface{}{
		{{"id; DROP", "=", 1}},
		{{"1=1--", "=", "x"}},
		{{"name", "REGEXP", ".*"}},
		{{"name", "=", "ok", "extra"}},
	}
	for _, dom := range badDomains {
		_, _, err := orm.BuildWhereWithRecordRules(ctx, 1, "test.safety", "read", dom)
		if err == nil {
			t.Fatalf("expected error for domain %v", dom)
		}
	}

	where, args, err := orm.BuildWhereWithRecordRules(ctx, 1, "test.safety", "read",
		[][]interface{}{{"name", "=", "alice"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(where, `$1`) {
		t.Fatalf("expected placeholder in WHERE, got %q", where)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("args=%v", args)
	}
}

func TestParseOrderByForModel_rejectsInjection(t *testing.T) {
	registerSafetyModel(t)
	bad := []string{"name; DROP", "id ASC; DELETE", "unknown_field"}
	for _, ob := range bad {
		if _, err := orm.ParseOrderByForModel("test.safety", ob); err == nil {
			t.Fatalf("expected reject order by %q", ob)
		}
	}
}

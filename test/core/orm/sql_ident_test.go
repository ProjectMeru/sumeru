package orm_test

import (
	"strings"
	"testing"
	"sumeru/core/orm"
)


func TestValidateModelName(t *testing.T) {
	ok := []string{"base", "contacts", "core.user", "sys.module", "mail.message", "sys.config.parameter"}
	for _, n := range ok {
		if err := orm.ValidateModelName(n); err != nil {
			t.Fatalf("%q: %v", n, err)
		}
	}
	bad := []string{"", "core_user", "core .user", "Core.User", "core..user", "1bad", "sys/rule", "sys;drop"}
	for _, n := range bad {
		if err := orm.ValidateModelName(n); err == nil {
			t.Fatalf("expected error for %q", n)
		}
	}
}

func TestModelToTableName(t *testing.T) {
	got, err := orm.ModelToTableName("core.user")
	if err != nil || got != "core_user" {
		t.Fatalf("got %q %v", got, err)
	}
	q, err := orm.QuotedTableName("core.user")
	if err != nil || q != `"core_user"` {
		t.Fatalf("quoted %q %v", q, err)
	}
}

func TestQuotedVsPhysical_indexNameMustNotEmbedQuotes(t *testing.T) {
	// MustQuotedTableName is for SQL clauses only. Index names must use the physical name
	// (MustModelToTableName); splicing the quoted form yields idx_"mail_message"_… which is invalid.
	physical := orm.MustModelToTableName("mail.message")
	if physical != "mail_message" {
		t.Fatalf("physical=%q", physical)
	}
	quoted := orm.MustQuotedTableName("mail.message")
	if quoted != `"mail_message"` {
		t.Fatalf("quoted=%q", quoted)
	}
	idx := "idx_" + physical + "_model_core_created"
	if strings.Contains(idx, `"`) {
		t.Fatalf("index name must not contain quotes: %q", idx)
	}
	if strings.Contains(idx, quoted) {
		t.Fatalf("index name must not embed quoted table: %q", idx)
	}
}

type identModel struct {
	name   string
	fields []orm.FieldDefinition
}

func (m identModel) ModelName() string              { return m.name }
func (m identModel) Fields() []orm.FieldDefinition { return m.fields }

func TestQuotedColumnForModel(t *testing.T) {
	m := identModel{
		name:   "test.ident",
		fields: []orm.FieldDefinition{{Name: "name", Type: orm.Char}},
	}
	orm.RegisterModelWithModule(m, "test")
	defer delete(orm.Registry, "test.ident")

	col, err := orm.QuotedColumnForModel("test.ident", "name")
	if err != nil || col != `"name"` {
		t.Fatalf("name: %q %v", col, err)
	}
	_, err = orm.QuotedColumnForModel("test.ident", "id; drop")
	if err == nil {
		t.Fatal("expected reject injection")
	}
	_, err = orm.QuotedColumnForModel("test.ident", "name;drop")
	if err == nil {
		t.Fatal("expected reject name;drop")
	}
	_, err = orm.QuotedColumnForModel("test.ident", "unknown")
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field, got %v", err)
	}
}

func TestParseOrderByForModel(t *testing.T) {
	m := identModel{
		name:   "test.order",
		fields: []orm.FieldDefinition{{Name: "name", Type: orm.Char}},
	}
	orm.RegisterModelWithModule(m, "test")
	defer delete(orm.Registry, "test.order")

	got, err := orm.ParseOrderByForModel("test.order", "name DESC")
	if err != nil || got != `"name" DESC` {
		t.Fatalf("got %q %v", got, err)
	}
	_, err = orm.ParseOrderByForModel("test.order", "id; select 1")
	if err == nil {
		t.Fatal("expected reject free-form order")
	}
	def, err := orm.ParseOrderByForModel("test.order", "")
	if err != nil || def != `"id" ASC` {
		t.Fatalf("default %q %v", def, err)
	}
}

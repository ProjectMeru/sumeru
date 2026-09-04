package orm_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/orm"
)

type stubUserModel struct{}

func (stubUserModel) ModelName() string { return "core.user" }
func (stubUserModel) Fields() []orm.FieldDefinition {
	return []orm.FieldDefinition{
		{Name: "id", Type: orm.Integer},
		{Name: "login", Type: orm.Char},
		{Name: "password", Type: orm.Char},
	}
}

func TestPrepareValuesRejectsDirectPassword(t *testing.T) {
	t.Parallel()
	_, err := orm.PrepareValues(stubUserModel{}, map[string]interface{}{
		"login":    "a@b.c",
		"password": "plaintext",
	}, orm.WriteOpWrite, orm.PrepareOptions{})
	if err == nil || !strings.Contains(err.Error(), "password cannot be set directly") {
		t.Fatalf("want direct password rejection, got %v", err)
	}
}

func TestPrepareValuesAllowsPasswordHashOption(t *testing.T) {
	t.Parallel()
	out, err := orm.PrepareValues(stubUserModel{}, map[string]interface{}{
		"password": "$2a$10$abcdefghijklmnopqrstuv",
	}, orm.WriteOpWrite, orm.PrepareOptions{AllowPasswordHash: true})
	if err != nil {
		t.Fatal(err)
	}
	if out["password"] == nil {
		t.Fatal("expected password in prepared values")
	}
}

func TestSetUserPasswordRequiresAdmin(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 2)
	err := orm.SetUserPassword(ctx, 2, 2, "ValidPass1")
	if err == nil {
		t.Fatal("expected denial without system admin (and/or DB)")
	}
}

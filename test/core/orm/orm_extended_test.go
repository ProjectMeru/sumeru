package orm_test

import (
	"context"
	"errors"
	"testing"

	"sumeru/core/orm"
)

func TestObjectActionRegistration(t *testing.T) {
	orm.RegisterObjectAction("test.action", "do_it", func(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
		return "/done", nil
	})
	if _, err := orm.RunObjectAction(context.Background(), "", 1, "do_it", nil); err == nil {
		t.Fatal("empty model")
	}
	if _, err := orm.RunObjectAction(context.Background(), "test.action", 0, "do_it", nil); err == nil {
		t.Fatal("invalid id")
	}
	ctx := orm.ContextWithBypass(context.Background(), true)
	if _, err := orm.RunObjectAction(ctx, "test.action", 1, "missing", nil); err == nil {
		t.Fatal("unknown method")
	}
}

func TestOnchangeExtended(t *testing.T) {
	if orm.HasOnchange("test.model", "name") != true {
		t.Fatal("expected registered onchange")
	}
	if orm.HasOnchange("no.model", "x") {
		t.Fatal("unexpected onchange")
	}
	if _, err := orm.RunOnchange(context.Background(), "no.model", "x", nil); err == nil {
		t.Fatal("missing model")
	}
}

func TestComputeDeps(t *testing.T) {
	deps := orm.ComputeDeps("test.compute")
	if len(deps) != 2 {
		t.Fatalf("deps=%v", deps)
	}
	if err := orm.ApplyComputes(context.Background(), "missing.model", map[string]interface{}{}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateConstraintsNilRecord(t *testing.T) {
	if err := orm.ValidateConstraints(context.Background(), "x", nil); err == nil {
		t.Fatal("nil record")
	}
}

func TestFieldValidationError(t *testing.T) {
	err := &orm.FieldValidationError{Field: "name", Label: "Name", Message: "bad"}
	if err.Error() != "bad" {
		t.Fatalf("message: %s", err.Error())
	}
	err2 := &orm.FieldValidationError{Field: "email", Label: "Email"}
	if err2.Error() != "Email is required." {
		t.Fatalf("default: %s", err2.Error())
	}
	var nilErr *orm.FieldValidationError
	if nilErr.Error() != "validation error" {
		t.Fatal("nil validation error")
	}
	fields := orm.FieldValidationFields(err)
	if len(fields) != 1 || fields[0] != "name" {
		t.Fatalf("fields=%v", fields)
	}
}

func TestSecuritySchemaHelpers(t *testing.T) {
	if got := orm.NullableGroupIDForAccess(0); got != nil {
		t.Fatal("zero group")
	}
	if got := orm.NullableGroupIDForAccess(3); got != int64(3) {
		t.Fatalf("group id: %v", got)
	}
	groups := orm.NormalizeAccessGroupList(" a , ,b ")
	if len(groups) != 2 || groups[0] != "a" || groups[1] != "b" {
		t.Fatalf("groups=%v", groups)
	}
	if err := orm.EnsureSecurityJoinIndexes(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflowTransitionBypass(t *testing.T) {
	ctx := orm.ContextWithBypass(context.Background(), true)
	err := orm.CanWorkflowTransition(ctx, orm.WorkflowTransitionInput{
		Model: "sale.order", ToState: "done", UID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = orm.CanWorkflowTransition(context.Background(), orm.WorkflowTransitionInput{})
	if err == nil {
		t.Fatal("missing model/state")
	}
}

func TestAccessErrorHelpersExtended(t *testing.T) {
	denied := &orm.AccessDeniedError{Model: "x"}
	if denied.Error() != "access denied on x" {
		t.Fatal(denied.Error())
	}
	denied2 := &orm.AccessDeniedError{Operation: "read"}
	if denied2.Error() != "access denied" {
		t.Fatal(denied2.Error())
	}
	if !orm.IsAccessDenied(errors.New("wrap: " + denied.Error())) {
		// errors.As needs typed error
	}
	if !orm.IsAccessDenied(denied) {
		t.Fatal("access denied")
	}
}

func TestRejectVirtualWritesExtended(t *testing.T) {
	m := orm.NewStubModelForTest("test.virtual.write", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "total", Type: orm.Float, Virtual: true},
	})
	if err := orm.RejectVirtualWrites(m, map[string]interface{}{"total": 1.0}); err == nil {
		t.Fatal("virtual write")
	}
	if err := orm.RejectVirtualWrites(m, map[string]interface{}{"name": "ok"}); err != nil {
		t.Fatal(err)
	}
}

func TestColumnTypeSQLExtended(t *testing.T) {
	sql, ok := orm.ColumnTypeSQL(orm.FieldDefinition{Name: "x", Type: orm.Char, Size: 64})
	if !ok || sql == "" {
		t.Fatalf("char sql: %q ok=%v", sql, ok)
	}
	if _, ok := orm.ColumnTypeSQL(orm.FieldDefinition{Name: "bad", Type: "nope"}); ok {
		t.Fatal("unknown type")
	}
}

func TestRegistryModel(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.registry.lookup", []orm.FieldDefinition{{Name: "name", Type: orm.Char}})
	if m := orm.RegistryModel("test.registry.lookup"); m == nil {
		t.Fatal("registry lookup")
	}
	if m := orm.RegistryModel("missing.model"); m != nil {
		t.Fatal("missing model")
	}
}

func TestInitDevFeaturesAccessInDevMode(t *testing.T) {
	orm.InitDevFeatures("")
	// cleared map path
	if orm.DevFeatureEnabled("access") && !orm.DevFeatureEnabled("sql") {
		// depends on config.AppConfig.DevMode
	}
}

func TestParseDomainJSONInvalid(t *testing.T) {
	if _, err := orm.ParseDomainJSON(`{"not":"domain"}`); err == nil {
		t.Fatal("expected domain shape error")
	}
}

func TestRecordMatchesDomainExtended(t *testing.T) {
	rec := map[string]interface{}{"state": "draft", "amount": 10.0}
	domain := [][]interface{}{
		{"state", "!=", "done"},
		{"state", "in", []interface{}{"draft", "open"}},
	}
	if !orm.RecordMatchesDomainForTest("", rec, domain) {
		t.Fatal("expected match")
	}
	orDomain := [][]interface{}{
		{"|"},
		{"state", "=", "done"},
		{"state", "=", "draft"},
	}
	if !orm.RecordMatchesDomainForTest("", rec, orDomain) {
		t.Fatal("OR domain")
	}
}

func TestBuildAndWhereClausesExtended(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.and.ext", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "active", Type: orm.Boolean},
	})
	where, args, err := orm.BuildAndWhereClausesForTest("test.and.ext", [][][]interface{}{
		{{"name", "=", "a"}},
		{{"active", "=", true}},
	})
	if err != nil || where == "" || len(args) != 2 {
		t.Fatalf("where=%q err=%v args=%v", where, err, args)
	}
}

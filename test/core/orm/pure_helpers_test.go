package orm_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/orm"
)

func TestAccessErrors(t *testing.T) {
	denied := &orm.AccessDeniedError{Model: "core.user", Operation: "read"}
	if denied.Error() != "access denied on core.user for operation read" {
		t.Fatalf("denied: %s", denied.Error())
	}
	if !orm.IsAccessDenied(denied) || orm.IsRecordRuleFailed(denied) {
		t.Fatal("IsAccessDenied mismatch")
	}
	rule := &orm.RecordRuleError{Model: "sale.order"}
	if rule.Error() != "record rule failed for model sale.order" {
		t.Fatalf("rule: %s", rule.Error())
	}
	if !orm.IsRecordRuleFailed(rule) {
		t.Fatal("IsRecordRuleFailed")
	}
	var nilDenied *orm.AccessDeniedError
	if nilDenied.Error() != "access denied" {
		t.Fatal("nil denied error")
	}
	var nilRule *orm.RecordRuleError
	if nilRule.Error() != "record rule failed" {
		t.Fatal("nil rule error")
	}
}

func TestSecurityContext(t *testing.T) {
	ctx := context.Background()
	if orm.UIDFromContext(nil) != 0 || orm.BypassFromContext(nil) {
		t.Fatal("nil context")
	}
	ctx = orm.ContextWithUID(ctx, 7)
	if orm.UIDFromContext(ctx) != 7 || orm.SecurityUID(ctx) != 7 {
		t.Fatal("uid")
	}
	ctx = orm.ContextWithCompanyID(ctx, 0)
	if orm.CompanyIDFromContext(ctx) != 0 {
		t.Fatal("zero company ignored")
	}
	ctx = orm.ContextWithCompanyID(ctx, 42)
	if orm.CompanyIDFromContext(ctx) != 42 {
		t.Fatal("company id")
	}
	ctx = orm.ContextWithBypass(ctx, true)
	if !orm.BypassFromContext(ctx) || !orm.SecurityBypass(ctx) {
		t.Fatal("bypass")
	}
}

func TestValidateModelAndFieldNames(t *testing.T) {
	cases := []struct {
		name    string
		fn      func(string) error
		valid   []string
		invalid []string
	}{
		{
			name:    "model",
			fn:      orm.ValidateModelName,
			valid:   []string{"core.user", "sale.order.line"},
			invalid: []string{"", "bad_name", "core..user", "Core.User", "core user"},
		},
		{
			name:    "field",
			fn:      orm.ValidateFieldName,
			valid:   []string{"name", "partner_id", "x2many_ids"},
			invalid: []string{"", "BadField", "bad-field"},
		},
	}
	for _, tc := range cases {
		for _, v := range tc.valid {
			if err := tc.fn(v); err != nil {
				t.Errorf("%s valid %q: %v", tc.name, v, err)
			}
		}
		for _, v := range tc.invalid {
			if err := tc.fn(v); err == nil {
				t.Errorf("%s invalid %q: expected error", tc.name, v)
			}
		}
	}
	if got, err := orm.ModelToTableName("core.user"); err != nil || got != "core_user" {
		t.Fatalf("ModelToTableName: %q err=%v", got, err)
	}
	if got := orm.MustModelToTableName("core.user"); got != "core_user" {
		t.Fatalf("MustModelToTableName: %q", got)
	}
	quoted, err := orm.QuotedTableName("core.user")
	if err != nil || !strings.Contains(quoted, "core_user") {
		t.Fatalf("QuotedTableName: %q err=%v", quoted, err)
	}
	if q := orm.MustQuotedTableName("core.user"); !strings.Contains(q, "core_user") {
		t.Fatalf("MustQuotedTableName: %q", q)
	}
}

func TestDevFeatures(t *testing.T) {
	orm.InitDevFeatures("sql,access,xml")
	for _, feat := range []string{"sql", "access", "xml"} {
		if !orm.DevFeatureEnabled(feat) {
			t.Fatalf("expected %q enabled", feat)
		}
	}
	if orm.DevFeatureEnabled("missing") {
		t.Fatal("unexpected feature")
	}
	orm.InitDevFeatures("")
	if orm.DevFeatureEnabled("SQL") {
		t.Fatal("cleared features should not match SQL unless dev mode")
	}
}

func TestBuildListSearchDomain(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.search", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "note", Type: orm.Text},
		{Name: "active", Type: orm.Boolean},
	})
	if got := orm.BuildListSearchDomain("test.search", nil, "x"); got != nil {
		t.Fatalf("empty fields: %v", got)
	}
	domain := orm.BuildListSearchDomain("test.search", []string{"name", "note", "active", "id"}, "acme")
	if len(domain) != 3 {
		t.Fatalf("expected OR domain, got %v", domain)
	}
	if domain[0][0] != "|" {
		t.Fatalf("expected OR prefix, got %v", domain)
	}
	single := orm.BuildListSearchDomain("test.search", []string{"name"}, "solo")
	if len(single) != 1 {
		t.Fatalf("single field: %v", single)
	}
}

func TestMergeDomains(t *testing.T) {
	base := [][]interface{}{{"active", "=", true}}
	extra := [][]interface{}{{"name", "ilike", "%x%"}}
	merged := orm.MergeDomains(base, extra)
	if len(merged) != 2 {
		t.Fatalf("merged: %v", merged)
	}
	if got := orm.MergeDomains(base, nil); len(got) != 1 {
		t.Fatalf("base only: %v", got)
	}
	if got := orm.MergeDomains(nil, extra); len(got) != 1 {
		t.Fatalf("extra only: %v", got)
	}
}

func TestParseDomainJSONAndSubstitute(t *testing.T) {
	domain, err := orm.ParseDomainJSON(`[["name", "=", "x"]]`)
	if err != nil || len(domain) != 1 {
		t.Fatalf("ParseDomainJSON: %v err=%v", domain, err)
	}
	if _, err := orm.ParseDomainJSON("not-json"); err == nil {
		t.Fatal("expected parse error")
	}
	sub := orm.SubstituteDomainUID([][]interface{}{{"user_id", "=", "$uid"}}, 5)
	if sub[0][2] != int64(5) {
		t.Fatalf("SubstituteDomainUID: %v", sub)
	}
	dc := orm.DomainContext{UID: 3, CompanyID: 1, CompanyIDs: []int64{1, 2}}
	out := orm.SubstituteDomainContext([][]interface{}{{"company_id", "=", "$company_id"}}, dc)
	if out[0][2] != int64(1) {
		t.Fatalf("SubstituteDomainContext company: %v", out)
	}
}

func TestAsBool(t *testing.T) {
	tests := map[interface{}]bool{
		true: true, false: false, 1: true, 0: false, "true": true, "1": true, "": false, nil: false,
	}
	for in, want := range tests {
		if got := orm.AsBool(in); got != want {
			t.Errorf("AsBool(%#v) = %v want %v", in, got, want)
		}
	}
}

func TestParseX2MCommandsExtended(t *testing.T) {
	cmds, err := orm.ParseX2MCommands([]interface{}{
		[]interface{}{int64(1), int64(2), map[string]interface{}{"name": "upd"}},
		[]interface{}{int64(2), int64(3)},
		[]interface{}{int64(0), int64(0), map[string]interface{}{"name": "new"}},
		[]interface{}{int64(4), int64(9)},
		[]interface{}{int64(5)},
		[]interface{}{int64(6), []interface{}{int64(1), int64(2)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 6 {
		t.Fatalf("cmds: %v", cmds)
	}
	if _, err := orm.ParseX2MCommands([]interface{}{"bad"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestTranslationCSVParse(t *testing.T) {
	csv := "lang,src,value,module\nen,Hello,Hi,base\n"
	header, rows, err := orm.ParseTranslationCSV(strings.NewReader(csv))
	if err != nil || len(rows) != 1 || len(header) != 4 {
		t.Fatalf("parse: header=%v rows=%v err=%v", header, rows, err)
	}
	if _, _, err := orm.ParseTranslationCSV(strings.NewReader("bad\nrow")); err == nil {
		t.Fatal("expected csv error")
	}
	if _, _, err := orm.ParseTranslationCSV(strings.NewReader("lang,src\nen,Hi")); err == nil {
		t.Fatal("expected missing column")
	}
	table, err := orm.TranslationTableName()
	if err != nil || table != "sys_translation" {
		t.Fatalf("table: %q err=%v", table, err)
	}
}

func TestCriteriaToDomain(t *testing.T) {
	domain := orm.CriteriaToDomain(map[string]interface{}{"name": "x", "active": true})
	if len(domain) != 2 {
		t.Fatalf("domain: %v", domain)
	}
	if got := orm.CriteriaToDomain(nil); len(got) != 0 {
		t.Fatalf("nil criteria: %v", got)
	}
}

func TestIsVirtualFieldAndFieldDef(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.virtual", []orm.FieldDefinition{
		{Name: "name", Type: orm.Char},
		{Name: "total", Type: orm.Float, Virtual: true},
	})
	if orm.IsVirtualField(orm.FieldDefinition{Name: "x", Virtual: true}) != true {
		t.Fatal("virtual")
	}
	if fd := orm.FieldDef("test.virtual", "name"); fd == nil || fd.Name != "name" {
		t.Fatalf("FieldDef: %#v", fd)
	}
	if fd := orm.FieldDef("test.virtual", "missing"); fd != nil {
		t.Fatal("missing field")
	}
}

func TestModelModuleHelpers(t *testing.T) {
	if !orm.IsPlatformModule("base") {
		t.Fatal("base is platform")
	}
	orm.RegisterStubModelForTest(t, "test.mod", []orm.FieldDefinition{{Name: "name", Type: orm.Char}})
	orm.RecordModelExtendedBy("test.mod", "custom")
	if mod := orm.DeclaringModule("test.mod"); mod == "" {
		t.Fatal("declaring module")
	}
	names := orm.ModelsExtendedByModule("custom")
	if len(names) != 1 || names[0] != "test.mod" {
		t.Fatalf("extended: %v", names)
	}
}

func TestHashAPIKey(t *testing.T) {
	h1 := orm.HashAPIKey("secret")
	h2 := orm.HashAPIKey("secret")
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash: %q %q", h1, h2)
	}
	raw, prefix, hash, err := orm.GenerateAPIKey()
	if err != nil || raw == "" || prefix == "" || hash == "" {
		t.Fatalf("GenerateAPIKey: raw=%q prefix=%q hash=%q err=%v", raw, prefix, hash, err)
	}
}

package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func TestModelRequiresCompanyIsolation(t *testing.T) {
	orm.RegisterStubModelForTest(t, "test.co.isolated", []orm.FieldDefinition{
		{Name: "company_id", Type: orm.Many2One, Relation: "core.company"},
	})
	if !orm.ModelRequiresCompanyIsolation("test.co.isolated") {
		t.Fatal("expected company isolation")
	}
	if orm.ModelRequiresCompanyIsolation("sys.module") {
		t.Fatal("sys models are global")
	}
	orm.SetModelCompanyShared("test.co.isolated", true)
	if orm.ModelRequiresCompanyIsolation("test.co.isolated") {
		t.Fatal("shared tag should skip isolation")
	}
	orm.SetModelCompanyShared("test.co.isolated", false)
}

func TestCompanyIsolationDomain_crossCompanyDenied(t *testing.T) {
	dom := orm.SubstituteDomainContext(orm.CompanyIsolationDomain(), orm.DomainContext{
		CompanyIDs: []int64{1},
	})
	otherCo := map[string]interface{}{"company_id": int64(2)}
	if orm.RecordMatchesDomainForTest("m", otherCo, dom) {
		t.Fatal("company 2 must not match allowed set [1]")
	}
	sameCo := map[string]interface{}{"company_id": int64(1)}
	if !orm.RecordMatchesDomainForTest("m", sameCo, dom) {
		t.Fatal("company 1 should match")
	}
	shared := map[string]interface{}{"company_id": false}
	if !orm.RecordMatchesDomainForTest("m", shared, dom) {
		t.Fatal("company_id false (shared) should match")
	}
}

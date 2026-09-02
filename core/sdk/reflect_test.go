package sdk_test

import (
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
)

type testPriority string

const (
	testPriorityLow testPriority = "low"
)

type testLine struct {
	modelmeta.ModelMeta `sumeru:"model=test.line"`
	Name                modelmeta.String
}

type corePartner struct {
	modelmeta.ModelMeta `sumeru:"model=core.partner"`
}

type testOrder struct {
	modelmeta.ModelMeta `sumeru:"model=test.order"`
	Name                modelmeta.String
	Amount              modelmeta.Float
	Rate                modelmeta.Float64
	State               modelmeta.String `sumeru:"selection=low:Low,high:High,default=low"`
	LineIds             modelmeta.One2Many[testLine]
	PartnerID           modelmeta.Many2One[corePartner] `sumeru:"string=Partner"`
}

type cookbookPriority string

const (
	cookbookPriorityLow  cookbookPriority = "low"
	cookbookPriorityHigh cookbookPriority = "high"
)

type cookbookLine struct {
	modelmeta.ModelMeta `sumeru:"model=test.cookbook.line"`
	Name                modelmeta.String
}

type cookbookOrder struct {
	modelmeta.ModelMeta `sumeru:"model=test.cookbook"`
	Name                modelmeta.String
	Priority            modelmeta.Selection[cookbookPriority] `sumeru:"default=low,selection=low:Low,high:High"`
	Rating              modelmeta.Integer                     `sumeru:"min=0,max=5"`
	State               modelmeta.String                      `sumeru:"selection=open:Open,closed:Closed"`
	CompanyName         modelmeta.String                      `sumeru:"related=partner_id.name"`
	LineIds             modelmeta.One2Many[cookbookLine]
}

func TestMustRegisterReflectsFields(t *testing.T) {
	orm.Registry = map[string]orm.Model{}
	modelreg.MustRegister("test", &testOrder{}, &testLine{})
	m, ok := orm.Registry["test.order"]
	if !ok {
		t.Fatal("test.order not registered")
	}
	fields := map[string]orm.FieldDefinition{}
	for _, f := range m.Fields() {
		fields[f.Name] = f
	}
	if fields["amount"].Type != orm.Float {
		t.Fatalf("amount type: got %s want float", fields["amount"].Type)
	}
	if fields["rate"].Type != orm.Float64 {
		t.Fatalf("rate type: got %s want float64", fields["rate"].Type)
	}
	if fields["state"].Type != orm.Selection {
		t.Fatalf("state type: got %s want selection", fields["state"].Type)
	}
	if fields["state"].DefaultVal != "low" {
		t.Fatalf("state default: got %v want low", fields["state"].DefaultVal)
	}
	if len(fields["state"].Selection) != 2 {
		t.Fatalf("state options: got %v", fields["state"].Selection)
	}
	if fields["partner_id"].Relation != "core.partner" {
		t.Fatalf("partner relation: %q", fields["partner_id"].Relation)
	}
	if fields["partner_id"].Relation != "core.partner" {
		t.Fatalf("partner relation: %q", fields["partner_id"].Relation)
	}
}

func TestFieldNameFromGo(t *testing.T) {
	if got := modelmeta.FieldNameFromGo("CompanyID"); got != "company_id" {
		t.Fatalf("got %q", got)
	}
}

func TestSelectionTypeRegistration(t *testing.T) {
	orm.Registry = map[string]orm.Model{}
	modelreg.MustRegister("test", &cookbookOrder{}, &cookbookLine{})
	m, ok := orm.Registry["test.cookbook"]
	if !ok {
		t.Fatal("test.cookbook not registered")
	}
	fields := map[string]orm.FieldDefinition{}
	for _, f := range m.Fields() {
		fields[f.Name] = f
	}
	if fields["priority"].Type != orm.Selection {
		t.Fatalf("priority type: got %s want selection", fields["priority"].Type)
	}
	if len(fields["priority"].Selection) != 2 {
		t.Fatalf("priority options: got %v", fields["priority"].Selection)
	}
	if fields["rating"].Min == nil || *fields["rating"].Min != 0 {
		t.Fatalf("rating min: %+v", fields["rating"].Min)
	}
	if fields["rating"].Max == nil || *fields["rating"].Max != 5 {
		t.Fatalf("rating max: %+v", fields["rating"].Max)
	}
	if fields["company_name"].Related != "partner_id.name" {
		t.Fatalf("related: %q", fields["company_name"].Related)
	}
	if !fields["company_name"].Virtual {
		t.Fatal("related field should be virtual")
	}
}

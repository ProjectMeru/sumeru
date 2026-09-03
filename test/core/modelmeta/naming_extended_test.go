package modelmeta_test

import (
	"reflect"
	"testing"

	"sumeru/core/modelmeta"
)

func TestLabelFromGo(t *testing.T) {
	if got := modelmeta.LabelFromGo("PartnerName"); got != "Partner Name" {
		t.Fatalf("got %q", got)
	}
	if got := modelmeta.LabelFromGo(""); got != "" {
		t.Fatalf("empty: %q", got)
	}
}

func TestModelNameFromGo(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"ResPartner":    "res.partner",
		"SaleOrderLine": "sale.order.line",
	}
	for in, want := range cases {
		if got := modelmeta.ModelNameFromGo(in); got != want {
			t.Errorf("ModelNameFromGo(%q) = %q want %q", in, got, want)
		}
	}
}

func TestHeuristicGoName(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"-":             "",
		"core.user":     "CoreUser",
		"sale.order":    "SaleOrder",
		"res.partner":   "ResPartner",
	}
	for in, want := range cases {
		if got := modelmeta.HeuristicGoName(in); got != want {
			t.Errorf("HeuristicGoName(%q) = %q want %q", in, got, want)
		}
	}
}

func TestParseFieldTagAllOptions(t *testing.T) {
	tag := "required,unique,index,readonly,store,default=x,column=col,size=10,precision=2,scale=3," +
		"string=Label,comodel=res.partner,ondelete=cascade,inverse=line_ids,table=t,left=l,right=r," +
		"model_field=mf,min=1,max=9,help=Help,currency=USD,groups=base.group_user," +
		"related=partner_id.name,compute=_compute_x"
	tags, err := modelmeta.ParseFieldTag(tag)
	if err != nil {
		t.Fatal(err)
	}
	if !tags.Required || !tags.Unique || !tags.Index || !tags.Readonly || !tags.Store {
		t.Fatal("flags")
	}
	if tags.Default != "x" || tags.Column != "col" || tags.Size != 10 || tags.Precision != 2 || tags.Scale != 3 {
		t.Fatalf("numeric tags: %+v", tags)
	}
	if tags.Label != "Label" || tags.Comodel != "res.partner" || tags.OnDelete != "cascade" {
		t.Fatalf("relation tags: %+v", tags)
	}
	if tags.Min == nil || *tags.Min != 1 || tags.Max == nil || *tags.Max != 9 {
		t.Fatalf("min/max: %+v", tags)
	}
	if tags.Help != "Help" || tags.Currency != "USD" || tags.Related != "partner_id.name" {
		t.Fatalf("misc: %+v", tags)
	}
}

func TestParseFieldTagRelationAlias(t *testing.T) {
	tags, err := modelmeta.ParseFieldTag("relation=res.company,label=Company")
	if err != nil || tags.Comodel != "res.company" || tags.Label != "Company" {
		t.Fatalf("relation alias: %+v err=%v", tags, err)
	}
}

func TestParseFieldTagInherits(t *testing.T) {
	tags, err := modelmeta.ParseFieldTag("inherits=res.partner")
	if err != nil || tags.Inherits != "res.partner" {
		t.Fatalf("inherits: %+v err=%v", tags, err)
	}
}

func TestModelSpecFromStruct(t *testing.T) {
	type SaleOrder struct {
		modelmeta.ModelMeta `sumeru:"model=sale.order"`
		Name                string
	}
	st := reflect.TypeOf(SaleOrder{})
	spec, err := modelmeta.ModelSpecFromStruct(st)
	if err != nil || spec.Name != "sale.order" {
		t.Fatalf("spec: %+v err=%v", spec, err)
	}
	name, err := modelmeta.ModelNameFromStruct(st)
	if err != nil || name != "sale.order" {
		t.Fatalf("name: %q err=%v", name, err)
	}
}

package modelmeta

import "testing"

func TestParseFieldTag(t *testing.T) {
	tags, err := ParseFieldTag("required,comodel=res.partner,string=Label")
	if err != nil {
		t.Fatalf("ParseFieldTag: %v", err)
	}
	if !tags.Required || tags.Comodel != "res.partner" || tags.Label != "Label" {
		t.Fatalf("unexpected tags: %+v", tags)
	}

	_, err = ParseFieldTag("unknown_flag=1")
	if err == nil {
		t.Fatal("expected error for unknown tag")
	}

	_, err = ParseFieldTag("size=not-a-number")
	if err == nil {
		t.Fatal("expected error for invalid size")
	}

	_, err = ParseFieldTag("min=not-a-float")
	if err == nil {
		t.Fatal("expected error for invalid min")
	}

	tags, err = ParseFieldTag("required,selection=a,b,c")
	if err != nil {
		t.Fatalf("selection with commas: %v", err)
	}
	if tags.Selection != "a,b,c" || !tags.Required {
		t.Fatalf("unexpected tags: %+v", tags)
	}

	tags, err = ParseFieldTag("selection=low:Low,high:High,default=low")
	if err != nil {
		t.Fatalf("default after selection options: %v", err)
	}
	if tags.Selection != "low:Low,high:High" || tags.Default != "low" {
		t.Fatalf("unexpected tags: %+v", tags)
	}

	tags, err = ParseFieldTag("default=planned,selection=planned:Planned,done:Done")
	if err != nil {
		t.Fatalf("default before selection: %v", err)
	}
	if tags.Default != "planned" || tags.Selection != "planned:Planned,done:Done" {
		t.Fatalf("unexpected tags: %+v", tags)
	}

	_, err = ParseFieldTag("size=10,size=20")
	if err == nil {
		t.Fatal("expected error for duplicate tag")
	}
}

func TestExtractDefaultFromSelectionTail(t *testing.T) {
	selection, defaultValue := extractDefaultFromSelectionTail("low:Low,high:High,default=low")
	if selection != "low:Low,high:High" || defaultValue != "low" {
		t.Fatalf("selection=%q default=%q", selection, defaultValue)
	}

	selection, defaultValue = extractDefaultFromSelectionTail("planned:Planned,done:Done")
	if selection != "planned:Planned,done:Done" || defaultValue != "" {
		t.Fatalf("selection=%q default=%q", selection, defaultValue)
	}
}

func TestPeelSelectionTag(t *testing.T) {
	body, selection := peelSelectionTag("required,selection=draft:Draft,done:Done")
	if body != "required" || selection != "draft:Draft,done:Done" {
		t.Fatalf("body=%q selection=%q", body, selection)
	}
}

func TestParseModelTag(t *testing.T) {
	name, err := ParseModelTag("model=core.partner")
	if err != nil || name != "core.partner" {
		t.Fatalf("model tag: name=%q err=%v", name, err)
	}

	name, err = ParseModelTag("inherit=core.partner")
	if err != nil || name != "core.partner" {
		t.Fatalf("inherit tag: name=%q err=%v", name, err)
	}

	name, err = ParseModelTag("model=-")
	if err != nil || name != "-" {
		t.Fatalf("skip tag: name=%q err=%v", name, err)
	}

	_, err = ParseModelTag("model=a,inherit=b")
	if err == nil {
		t.Fatal("expected exclusivity error")
	}
}

func TestModelSpecFromTags(t *testing.T) {
	spec, err := ModelSpecFromTags(FieldTags{Model: "sale.order"}, "SaleOrder")
	if err != nil || spec.Name != "sale.order" || spec.Extend {
		t.Fatalf("model spec: %+v err=%v", spec, err)
	}

	spec, err = ModelSpecFromTags(FieldTags{Inherit: "sale.order"}, "SaleOrderLine")
	if err != nil || spec.Name != "sale.order" || !spec.Extend {
		t.Fatalf("inherit spec: %+v err=%v", spec, err)
	}

	spec, err = ModelSpecFromTags(FieldTags{}, "ResPartner")
	if err != nil || spec.Name != "res.partner" || spec.Extend {
		t.Fatalf("default name: %+v err=%v", spec, err)
	}

	_, err = ModelSpecFromTags(FieldTags{Model: "a", Inherit: "b"}, "Bad")
	if err == nil {
		t.Fatal("expected exclusivity error")
	}
}

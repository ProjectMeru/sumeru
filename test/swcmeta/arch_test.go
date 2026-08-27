package swcmeta_test

import (
	"context"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
)

func TestSerializeViewListArch(t *testing.T) {
	arch := `<view type="list" model="sale.order"><field name="name"/><field name="amount_total"/></view>`
	view, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	out := swcmeta.SerializeView(view)
	if out.Type != "list" || out.Model != "sale.order" {
		t.Fatalf("arch meta: %+v", out)
	}
	if len(out.Fields) != 2 {
		t.Fatalf("fields: %d", len(out.Fields))
	}
}

func TestBuildWorkspacePayloadRedacts(t *testing.T) {
	ctx := context.Background()
	rec := swcmeta.ViewRecordInput{
		ResModel: "test.model",
		Record:   map[string]interface{}{"password": "secret", "name": "Acme"},
	}
	// Without registry/ACL this is a smoke test that redact path runs.
	payload := swcmeta.BuildWorkspacePayload(ctx, &parser.View{Model: "test.model", Type: "form"}, "form", rec, "1")
	if payload.Record == nil {
		t.Fatal("expected record")
	}
	_ = orm.RedactRecordForRead // link orm package
}

func TestBuildWorkspacePayloadDefaults(t *testing.T) {
	payload := swcmeta.BuildWorkspacePayload(
		context.Background(),
		&parser.View{Model: "crm.lead", Type: "kanban"},
		"kanban",
		swcmeta.ViewRecordInput{
			ResModel: "crm.lead",
			Defaults: map[string]interface{}{"type": "opportunity"},
		},
		"1",
	)
	if payload.Defaults["type"] != "opportunity" {
		t.Fatalf("defaults: %+v", payload.Defaults)
	}
}

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

func TestSerializeGanttMapCohortArch(t *testing.T) {
	gantt, err := parser.ParseViewFromArch(`<gantt date_start="date_start" date_stop="date_end"><field name="name"/></gantt>`)
	if err != nil {
		t.Fatal(err)
	}
	ganttOut := swcmeta.SerializeView(gantt)
	if ganttOut.Gantt == nil || ganttOut.Gantt.DateStart != "date_start" || ganttOut.Calendar != nil {
		t.Fatalf("gantt meta: %+v", ganttOut)
	}

	mp, err := parser.ParseViewFromArch(`<map latitude="lat" longitude="lng"><field name="name"/></map>`)
	if err != nil {
		t.Fatal(err)
	}
	mapOut := swcmeta.SerializeView(mp)
	if mapOut.Map == nil || mapOut.Map.Latitude != "lat" || mapOut.Map.Longitude != "lng" {
		t.Fatalf("map meta: %+v", mapOut)
	}

	cohort, err := parser.ParseViewFromArch(`<cohort date_start="create_date" interval="week" measure="amount"><field name="name"/></cohort>`)
	if err != nil {
		t.Fatal(err)
	}
	cohortOut := swcmeta.SerializeView(cohort)
	if cohortOut.Cohort == nil || cohortOut.Cohort.Interval != "week" || cohortOut.Cohort.Measure != "amount" {
		t.Fatalf("cohort meta: %+v", cohortOut)
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

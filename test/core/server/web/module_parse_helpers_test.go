package web_test

import (
	"sumeru/core/server/web"
	"testing"

	"sumeru/core/orm"
)

func TestParseModuleRowFields(t *testing.T) {
	row, ok := web.ParseModuleRow(map[string]interface{}{
		"name":         "sale",
		"display_name": "Sales",
		"author":       "Meru",
		"version":      "1.0",
		"description":  "Sales app",
		"state":        "installed",
		"application":  true,
		"active":       true,
	})
	if !ok {
		t.Fatal("expected ok")
	}
	if row.Name != "sale" || row.DisplayName != "Sales" || row.Author != "Meru" {
		t.Fatalf("unexpected row: %+v", row)
	}
	if !row.Application || !row.Active || row.State != "installed" {
		t.Fatalf("unexpected flags: %+v", row)
	}
}

func TestParseModuleRowMissingName(t *testing.T) {
	_, ok := web.ParseModuleRow(map[string]interface{}{
		"display_name": "No technical name",
	})
	if ok {
		t.Fatal("expected false when name is empty")
	}
}

func TestModuleDisplayNameFallback(t *testing.T) {
	if got := web.ModuleDisplayName("crm", orm.AsString(nil)); got != "crm" {
		t.Fatalf("got %q", got)
	}
}

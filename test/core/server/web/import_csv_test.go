package web_test

import (
	"reflect"
	"sumeru/core/server/web"
	"testing"
)

func TestCoerceCSVValue(t *testing.T) {
	tests := []struct {
		raw  string
		want interface{}
	}{
		{raw: "", want: ""},
		{raw: " TRUE ", want: true},
		{raw: "false", want: false},
		{raw: "42", want: int64(42)},
		{raw: "hello", want: "hello"},
	}
	for _, test := range tests {
		if got := web.CoerceCSVValue(test.raw); !reflect.DeepEqual(got, test.want) {
			t.Fatalf("web.CoerceCSVValue(%q) = %#v want %#v", test.raw, got, test.want)
		}
	}
}

func TestImportableRowValues(t *testing.T) {
	header := []string{"id", "name", "active", "unknown"}
	record := []string{"1", " Acme ", "true", "skip"}
	allowedFields := map[string]struct{}{
		"name":   {},
		"active": {},
	}

	got := web.ImportableRowValues(header, record, allowedFields)
	want := map[string]interface{}{
		"name":   "Acme",
		"active": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestIsImportableColumn(t *testing.T) {
	allowedFields := map[string]struct{}{"name": {}}
	if web.IsImportableColumn("id", allowedFields) {
		t.Fatal("id column should not be importable")
	}
	if !web.IsImportableColumn("name", allowedFields) {
		t.Fatal("name column should be importable")
	}
}

func TestImportCSVFlashMessage(t *testing.T) {
	if got := web.ImportCSVFlashMessage(5); got != "imported_5" {
		t.Fatalf("got %q want imported_5", got)
	}
}

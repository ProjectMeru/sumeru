package api

import "testing"

func TestProjectFields(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": 1, "name": "a", "secret": "x"},
		{"id": 2, "name": "b", "secret": "y"},
	}
	out := projectFields(rows, nil)
	if len(out) != 2 || out[0]["secret"] != "x" {
		t.Fatalf("empty fields should pass through: %+v", out)
	}
	out = projectFields(rows, []string{"id", "name"})
	if len(out) != 2 {
		t.Fatal("expected 2 rows")
	}
	if _, ok := out[0]["secret"]; ok {
		t.Fatalf("secret should be omitted: %+v", out[0])
	}
	if out[0]["id"] != 1 || out[0]["name"] != "a" {
		t.Fatalf("projected = %+v", out[0])
	}
}

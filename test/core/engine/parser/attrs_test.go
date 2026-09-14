package parser_test

import (
	"testing"

	"sumeru/core/engine/parser"
)

func TestParseXMLBoolAttr(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		raw  string
		want bool
	}{
		{"", false},
		{"true", true},
		{"FALSE", false},
	} {
		got, err := parser.ParseXMLBoolAttr("flag", tc.raw)
		if err != nil {
			t.Fatalf("ParseXMLBoolAttr(%q): %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("ParseXMLBoolAttr(%q) = %v; want %v", tc.raw, got, tc.want)
		}
	}
	for _, bad := range []string{"1", "0", "yes", "off"} {
		if _, err := parser.ParseXMLBoolAttr("flag", bad); err == nil {
			t.Fatalf("ParseXMLBoolAttr(%q) expected error", bad)
		}
	}
}

func TestParseModifierAttr(t *testing.T) {
	t.Parallel()
	lit, truthy, expr, err := parser.ParseModifierAttr("invisible", "true")
	if err != nil || !lit || !truthy || expr != "" {
		t.Fatalf("true: lit=%v truthy=%v expr=%q err=%v", lit, truthy, expr, err)
	}
	lit, truthy, expr, err = parser.ParseModifierAttr("invisible", "false")
	if err != nil || !lit || truthy || expr != "" {
		t.Fatalf("false: lit=%v truthy=%v expr=%q err=%v", lit, truthy, expr, err)
	}
	lit, truthy, expr, err = parser.ParseModifierAttr("invisible", "state != 'done'")
	if err != nil || lit || truthy || expr != "state != 'done'" {
		t.Fatalf("expr: lit=%v truthy=%v expr=%q err=%v", lit, truthy, expr, err)
	}
	if _, _, _, err := parser.ParseModifierAttr("invisible", "1"); err == nil {
		t.Fatal("expected error for invisible=1")
	}
}

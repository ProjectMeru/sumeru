package api

import (
	"encoding/json"
	"testing"
)

func TestNormArgs(t *testing.T) {
	if got := string(normArgs(nil)); got != "[]" {
		t.Fatalf("normArgs(nil) = %q", got)
	}
	if got := string(normArgs(json.RawMessage("null"))); got != "[]" {
		t.Fatalf("normArgs(null) = %q", got)
	}
	raw := json.RawMessage(`[1]`)
	if got := string(normArgs(raw)); got != "[1]" {
		t.Fatalf("normArgs([1]) = %q", got)
	}
}

func TestParseArgsArray(t *testing.T) {
	arr, err := parseArgsArray(nil)
	if err != nil || len(arr) != 0 {
		t.Fatalf("parseArgsArray(nil) = %v, %v", arr, err)
	}
	_, err = parseArgsArray(json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for non-array args")
	}
	if ce, ok := err.(*codedError); !ok || ce.code != CodeInvalidArgs {
		t.Fatalf("err = %v", err)
	}
	arr, err = parseArgsArray(json.RawMessage(`[1,2]`))
	if err != nil || len(arr) != 2 {
		t.Fatalf("parseArgsArray = %v, %v", arr, err)
	}
}

func TestParseDomainArg(t *testing.T) {
	d, err := parseDomainArg(nil)
	if err != nil || d != nil {
		t.Fatalf("parseDomainArg(nil) = %v, %v", d, err)
	}
	d, err = parseDomainArg(json.RawMessage(`[["id","=",1]]`))
	if err != nil || len(d) != 1 {
		t.Fatalf("parseDomainArg domain = %v, %v", d, err)
	}
	_, err = parseDomainArg(json.RawMessage(`"not-domain"`))
	if err == nil {
		t.Fatal("expected domain parse error")
	}
}

func TestParseLimitOffset(t *testing.T) {
	limit, offset := parseLimitOffset(nil)
	if limit != 500 || offset != 0 {
		t.Fatalf("defaults = %d, %d", limit, offset)
	}
	kw := json.RawMessage(`{"limit":10,"offset":5}`)
	limit, offset = parseLimitOffset(kw)
	if limit != 10 || offset != 5 {
		t.Fatalf("parsed = %d, %d", limit, offset)
	}
	kw = json.RawMessage(`{"limit":0,"offset":-1}`)
	limit, offset = parseLimitOffset(kw)
	if limit != 500 || offset != -1 {
		t.Fatalf("limit clamp only = %d, %d", limit, offset)
	}
}

func TestToFloat(t *testing.T) {
	if v, ok := toFloat(float64(3)); !ok || v != 3 {
		t.Fatalf("float64: %v, %v", v, ok)
	}
	if v, ok := toFloat(int(7)); !ok || v != 7 {
		t.Fatalf("int: %v, %v", v, ok)
	}
	if v, ok := toFloat(json.Number("2.5")); !ok || v != 2.5 {
		t.Fatalf("json.Number: %v, %v", v, ok)
	}
	if v, ok := toFloat(int64(7)); !ok || v != 7 {
		t.Fatalf("int64: %v, %v", v, ok)
	}
}

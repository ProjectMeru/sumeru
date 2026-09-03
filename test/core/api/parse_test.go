package api_test

import (
	"sumeru/core/server/api"
	"encoding/json"
	"testing"
)



func TestNormArgs(t *testing.T) {
	if got := string(api.NormArgsForTest(nil)); got != "[]" {
		t.Fatalf("api.NormArgsForTest(nil) = %q", got)
	}
	if got := string(api.NormArgsForTest(json.RawMessage("null"))); got != "[]" {
		t.Fatalf("api.NormArgsForTest(null) = %q", got)
	}
	raw := json.RawMessage(`[1]`)
	if got := string(api.NormArgsForTest(raw)); got != "[1]" {
		t.Fatalf("api.NormArgsForTest([1]) = %q", got)
	}
}

func TestParseArgsArray(t *testing.T) {
	arr, err := api.ParseArgsArrayForTest(nil)
	if err != nil || len(arr) != 0 {
		t.Fatalf("api.ParseArgsArrayForTest(nil) = %v, %v", arr, err)
	}
	_, err = api.ParseArgsArrayForTest(json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for non-array args")
	}
	if code, ok := api.RPCErrorCodeForTest(err); !ok || code != api.CodeInvalidArgs {
		t.Fatalf("err = %v", err)
	}
	arr, err = api.ParseArgsArrayForTest(json.RawMessage(`[1,2]`))
	if err != nil || len(arr) != 2 {
		t.Fatalf("parseArgsArray = %v, %v", arr, err)
	}
}

func TestParseDomainArg(t *testing.T) {
	d, err := api.ParseDomainArgForTest(nil)
	if err != nil || d != nil {
		t.Fatalf("api.ParseDomainArgForTest(nil) = %v, %v", d, err)
	}
	d, err = api.ParseDomainArgForTest(json.RawMessage(`[["id","=",1]]`))
	if err != nil || len(d) != 1 {
		t.Fatalf("parseDomainArg domain = %v, %v", d, err)
	}
	_, err = api.ParseDomainArgForTest(json.RawMessage(`"not-domain"`))
	if err == nil {
		t.Fatal("expected domain parse error")
	}
}

func TestParseLimitOffset(t *testing.T) {
	limit, offset := api.ParseLimitOffsetForTest(nil)
	if limit != 500 || offset != 0 {
		t.Fatalf("defaults = %d, %d", limit, offset)
	}
	kw := json.RawMessage(`{"limit":10,"offset":5}`)
	limit, offset = api.ParseLimitOffsetForTest(kw)
	if limit != 10 || offset != 5 {
		t.Fatalf("parsed = %d, %d", limit, offset)
	}
	kw = json.RawMessage(`{"limit":0,"offset":-1}`)
	limit, offset = api.ParseLimitOffsetForTest(kw)
	if limit != 500 || offset != -1 {
		t.Fatalf("limit clamp only = %d, %d", limit, offset)
	}
}

func TestToFloat(t *testing.T) {
	if v, ok := api.ToFloatForTest(float64(3)); !ok || v != 3 {
		t.Fatalf("float64: %v, %v", v, ok)
	}
	if v, ok := api.ToFloatForTest(int(7)); !ok || v != 7 {
		t.Fatalf("int: %v, %v", v, ok)
	}
	if v, ok := api.ToFloatForTest(json.Number("2.5")); !ok || v != 2.5 {
		t.Fatalf("json.Number: %v, %v", v, ok)
	}
	if v, ok := api.ToFloatForTest(int64(7)); !ok || v != 7 {
		t.Fatalf("int64: %v, %v", v, ok)
	}
}

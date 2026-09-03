package web_test

import (
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestAcceptsJSONContentType(t *testing.T) {
	if !web.AcceptsJSONContentType("") {
		t.Fatal("empty content type should be accepted")
	}
	if !web.AcceptsJSONContentType(" application/json; charset=utf-8 ") {
		t.Fatal("json content type should be accepted")
	}
	if web.AcceptsJSONContentType("text/plain") {
		t.Fatal("non-json content type should be rejected")
	}
}

func TestParseRPCCallMeta(t *testing.T) {
	modelName, methodName := web.ParseRPCCallMeta([]byte(`{"model":"core.user","method":"search","args":[]}`))
	if modelName != "core.user" || methodName != "search" {
		t.Fatalf("got model=%q method=%q", modelName, methodName)
	}

	modelName, methodName = web.ParseRPCCallMeta([]byte(`{`))
	if modelName != "" || methodName != "" {
		t.Fatalf("invalid json should return empty meta, got model=%q method=%q", modelName, methodName)
	}
}

func TestReadBoundedRequestBody(t *testing.T) {
	body := `{"model":"core.user"}`
	req := httptest.NewRequest("POST", web.TestAPIRPCRoute, strings.NewReader(body))
	got, ok := web.ReadBoundedRequestBody(req, web.TestMaxRPCBodyBytes)
	if !ok || string(got) != body {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestReadBoundedRequestBodyOverLimit(t *testing.T) {
	largeBody := strings.Repeat("a", int(web.TestMaxRPCBodyBytes)+1)
	req := httptest.NewRequest("POST", web.TestAPIRPCRoute, strings.NewReader(largeBody))
	got, ok := web.ReadBoundedRequestBody(req, web.TestMaxRPCBodyBytes)
	if !ok || int64(len(got)) != web.TestMaxRPCBodyBytes+1 {
		t.Fatalf("expected over-limit body to be read, got len=%d ok=%v", len(got), ok)
	}
}

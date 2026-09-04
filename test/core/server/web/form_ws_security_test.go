package web_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/web"
)

func TestParsePostFormRejectsOversizedBody(t *testing.T) {
	// 1 MiB limit + 1 byte
	body := bytes.Repeat([]byte("a"), (1<<20)+1)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	if web.ParsePostForm(rec, req) {
		t.Fatal("oversized form body should be rejected")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestParsePostFormAcceptsSmallBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("a=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	if !web.ParsePostForm(rec, req) {
		t.Fatal("small form body should parse")
	}
	if req.Form.Get("a") != "1" {
		t.Fatalf("form a=%q want 1", req.Form.Get("a"))
	}
}

func TestCheckSwcBusOriginRejectsCrossOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://app.example/ws", nil)
	req.Host = "app.example"
	req.Header.Set("Origin", "https://evil.example")
	if web.CheckSwcBusOrigin(req) {
		t.Fatal("cross-origin WS upgrade should be rejected")
	}
}

func TestCheckSwcBusOriginAllowsSameOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://app.example/ws", nil)
	req.Host = "app.example"
	req.Header.Set("Origin", "http://app.example")
	if !web.CheckSwcBusOrigin(req) {
		t.Fatal("same-origin WS should be allowed")
	}
}

func TestCheckSwcBusOriginAllowsMissingOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://app.example/ws", nil)
	req.Host = "app.example"
	if !web.CheckSwcBusOrigin(req) {
		t.Fatal("missing Origin should be allowed (non-browser clients)")
	}
}

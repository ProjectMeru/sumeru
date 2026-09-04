package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/web"
)

func TestAPIHealthHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	web.APIHealthHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
}

func TestAPIReadyHandlerWithoutDB(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
	web.APIReadyHandler(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d want 503 when DB nil", rec.Code)
	}
}

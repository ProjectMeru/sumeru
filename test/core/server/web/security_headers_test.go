package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestSecurityHeadersSet(t *testing.T) {
	prev := config.AppConfig.DevMode
	config.AppConfig.DevMode = false
	t.Cleanup(func() { config.AppConfig.DevMode = prev })

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := web.SecurityMiddleware(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	h.ServeHTTP(rec, req)
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
	if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Fatal("missing frame options")
	}
	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Fatal("missing HSTS when not in dev_mode")
	}
}

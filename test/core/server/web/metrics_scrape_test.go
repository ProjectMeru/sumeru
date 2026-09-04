package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestMetricsHandler_scrapeToken(t *testing.T) {
	prev := config.AppConfig.MetricsScrapeToken
	config.AppConfig.MetricsScrapeToken = "scrape-secret"
	t.Cleanup(func() { config.AppConfig.MetricsScrapeToken = prev })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	web.MetricsHandler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("without token status=%d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer scrape-secret")
	web.MetricsHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("with token status=%d body=%s", rec.Code, rec.Body.String())
	}
}

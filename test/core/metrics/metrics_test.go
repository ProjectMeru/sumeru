package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sumeru/core/metrics"
)

func TestIncAndObserve(t *testing.T) {
	metrics.Inc("test_counter")
	metrics.Inc("test_counter")
	metrics.ObserveDuration("test_latency", 250*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	metrics.Handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "test_counter 2") {
		t.Fatalf("counter missing: %s", body)
	}
	if !strings.Contains(body, "test_latency_sum") || !strings.Contains(body, "test_latency_count 1") {
		t.Fatalf("histogram missing: %s", body)
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()
	metrics.Handler(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d", rec.Code)
	}
}

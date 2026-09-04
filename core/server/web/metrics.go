package web

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"sumeru/core/metrics"
	"sumeru/core/server/config"
)

// MetricsHandler exposes Prometheus text metrics.
// Auth: Bearer metrics_scrape_token when configured, otherwise system-admin session.
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(config.AppConfig.MetricsScrapeToken)
	if token != "" {
		got := bearerToken(r.Header.Get(authHeader))
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		metrics.Handler(w, r)
		return
	}
	if !requireSystemAdmin(w, r, false) {
		return
	}
	metrics.Handler(w, r)
}

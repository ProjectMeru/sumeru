package web_test

import (
	"net/http/httptest"
	"sumeru/core/server/web"
	"testing"
	"time"
)

func TestClientIPFromForwardedHeader(t *testing.T) {
	req := httptest.NewRequest("POST", web.TestSetupInitRoute, nil)
	req.Header.Set(web.TestForwardedForHeader, "203.0.113.1, 198.51.100.2")
	if got := web.ClientIP(req); got != "203.0.113.1" {
		t.Fatalf("got %q want first forwarded IP", got)
	}
}

func TestIsLoopbackIP(t *testing.T) {
	if !web.IsLoopbackIP("127.0.0.1") {
		t.Fatal("127.0.0.1 should be loopback")
	}
	if web.IsLoopbackIP("203.0.113.1") {
		t.Fatal("public IP should not be loopback")
	}
}

func TestSetupTokenFromRequestPrefersHeader(t *testing.T) {
	req := httptest.NewRequest("POST", web.TestSetupInitRoute, nil)
	req.Header.Set(web.TestSetupTokenHeader, "header-token")
	if got := web.SetupTokenFromRequest(req, "body-token"); got != "header-token" {
		t.Fatalf("got %q want header token", got)
	}
}

func TestSetupTokenFromRequestBodyFallback(t *testing.T) {
	req := httptest.NewRequest("POST", web.TestSetupInitRoute, nil)
	if got := web.SetupTokenFromRequest(req, "body-token"); got != "body-token" {
		t.Fatalf("got %q want body token", got)
	}
}

func TestPruneSetupAttempts(t *testing.T) {
	now := time.Now()
	attempts := []time.Time{
		now.Add(-2 * web.TestSetupRateLimitWindow),
		now.Add(-30 * time.Second),
	}
	pruned := web.PruneSetupAttempts(attempts, now)
	if len(pruned) != 1 {
		t.Fatalf("got %d attempts want 1 recent attempt", len(pruned))
	}
}

func TestAllowSetupRateLimit(t *testing.T) {
	web.ResetSetupRateLimiterForTest()

	requestIP := "127.0.0.1"
	recorder := httptest.NewRecorder()
	for i := 0; i < web.TestSetupRateLimitMax; i++ {
		if !web.AllowSetupRateLimit(recorder, requestIP) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if web.AllowSetupRateLimit(recorder, requestIP) {
		t.Fatal("attempt over limit should be rejected")
	}
}

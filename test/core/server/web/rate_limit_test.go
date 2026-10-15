package web_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/api"
	"sumeru/core/server/web"
)

func TestEnforceRateLimit_RPCReturnsJSONEnvelope(t *testing.T) {
	web.ResetRateLimitForTest()
	web.SetRateLimitRPMForTest(1)
	t.Cleanup(func() {
		web.ResetRateLimitForTest()
		web.SetRateLimitRPMForTest(0)
	})

	body := `{"domain":[]}`
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, web.TestRPCURLForTest("core.user", "search"), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		web.ApplyDefaultRPCHeadersForTest(req)
		req.RemoteAddr = "203.0.113.10:1234"
		if !web.EnforceRateLimitForTest(rr, req) {
			if rr.Code != http.StatusTooManyRequests {
				t.Fatalf("attempt %d status = %d, want 429", i, rr.Code)
			}
			if i == 0 {
				t.Fatal("first request should not be rate limited")
			}
			var env struct {
				Success bool          `json:"success"`
				Error   *api.APIError `json:"error"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
				t.Fatalf("decode: %v body=%q", err, rr.Body.String())
			}
			if env.Success || env.Error == nil || env.Error.Code != api.CodeRateLimited {
				t.Fatalf("body = %s", rr.Body.String())
			}
			return
		}
	}
	t.Fatal("expected second request to be rate limited")
}

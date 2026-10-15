package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/api"
	"sumeru/core/server/web"
)

func TestRPCAuthLogin_validationErrors(t *testing.T) {
	t.Run("missing password", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, web.TestAuthLoginURLForTest(), strings.NewReader(`{"login":"u@example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		web.ApplyDefaultRPCHeadersForTest(req)
		web.RPCJSONHandler(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; want 400", rr.Code)
		}
		env := decodeRPCEnvelope(t, rr)
		if env.Success || env.Error == nil || env.Error.Code != api.CodeValidationError {
			t.Fatalf("error = %+v; want VALIDATION", env.Error)
		}
	})

	t.Run("rejects bearer on login", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, web.TestAuthLoginURLForTest(), strings.NewReader(`{"login":"u","password":"p"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk_test")
		web.ApplyDefaultRPCHeadersForTest(req)
		web.RPCJSONHandler(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d; want 400", rr.Code)
		}
	})

	t.Run("GET not allowed", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, web.TestAuthLoginURLForTest(), nil)
		web.ApplyDefaultRPCHeadersForTest(req)
		web.RPCJSONHandler(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d; want 405", rr.Code)
		}
	})
}

func TestRPCAuthLogin_invalidCredentialsEnvelope(t *testing.T) {
	web.ResetLoginLockoutForTest()
	t.Cleanup(web.ResetLoginLockoutForTest)

	rr := httptest.NewRecorder()
	body := `{"login":"no-such-user@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, web.TestAuthLoginURLForTest(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	web.ApplyDefaultRPCHeadersForTest(req)
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", rr.Code)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.Success {
		t.Fatal("success = true; want false")
	}
	if env.Error == nil || env.Error.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("error = %+v; want INVALID_CREDENTIALS", env.Error)
	}
	if env.Error.Type != "AuthenticationError" {
		t.Fatalf("error.type = %q; want AuthenticationError", env.Error.Type)
	}
	if strings.Contains(rr.Body.String(), "sk_") {
		t.Fatal("response must not contain api key material")
	}
}

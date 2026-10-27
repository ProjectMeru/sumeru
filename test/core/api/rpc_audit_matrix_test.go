package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/api"
	"sumeru/core/server/web"
)

func TestRPCJSONHandler_unauthenticatedEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, web.TestRPCURLForTest("core.user", "search"), strings.NewReader(`{"domain":[]}`))
	req.Header.Set("Content-Type", "application/json")
	web.ApplyDefaultRPCHeadersForTest(req)
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", rr.Code)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.Success || env.Error == nil || env.Error.Code != api.CodeUnauthorized {
		t.Fatalf("env = %+v", env)
	}
	if env.Error.Message == "" {
		t.Fatal("expected non-empty error message")
	}
}

func rpcSessionRequest(model, method, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, web.TestRPCURLForTest(model, method), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	web.ApplyDefaultRPCHeadersForTest(req)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: "rpc-audit-test-sid"})
	req.Header.Set("X-CSRF-Token", web.CSRFTokenForRequestForTest(req))
	return req
}

func TestRPCJSONHandler_methodNotAllowedEnvelope(t *testing.T) {
	web.SetTestSessionUserIDForTest(1)
	t.Cleanup(web.ResetTestSessionUserIDForTest)

	rr := httptest.NewRecorder()
	web.RPCJSONHandler(rr, rpcSessionRequest("core.user", "drop_database", `{"domain":[]}`))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", rr.Code)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.Error == nil || env.Error.Code != api.CodeMethodNotAllowed {
		t.Fatalf("env = %+v", env)
	}
}

func TestRPCJSONHandler_unknownModelEnvelope(t *testing.T) {
	web.SetTestSessionUserIDForTest(1)
	t.Cleanup(web.ResetTestSessionUserIDForTest)

	rr := httptest.NewRecorder()
	web.RPCJSONHandler(rr, rpcSessionRequest("evil.model", "search", `{"domain":[]}`))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", rr.Code)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.Error == nil || env.Error.Code != api.CodeModelNotFound {
		t.Fatalf("env = %+v", env)
	}
}

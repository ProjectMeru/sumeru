package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/api"
	"sumeru/core/server/web"
)

type rpcEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *api.APIError   `json:"error"`
}

func decodeRPCEnvelope(t *testing.T, rr *httptest.ResponseRecorder) rpcEnvelope {
	t.Helper()
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q; want application/json", ct)
	}
	var env rpcEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rr.Body.String(), err)
	}
	return env
}

func TestRPCJSONHandler_unauthenticatedJSONEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, web.TestRPCURLForTest("core.user", "search"), strings.NewReader(`{"domain":[]}`))
	req.Header.Set("Content-Type", "application/json")
	web.ApplyDefaultRPCHeadersForTest(req)
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusUnauthorized)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.Success || env.Error == nil || env.Error.Code != api.CodeUnauthorized {
		t.Fatalf("error = %+v; want UNAUTHORIZED", env.Error)
	}
}

func TestRPCJSONHandler_nonPostJSONEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, web.TestRPCURLForTest("core.user", "search"), nil)
	web.ApplyDefaultRPCHeadersForTest(req)
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.Success {
		t.Fatal("success = true; want false")
	}
	if env.Error == nil || env.Error.Code != api.CodeMethodNotAllowed {
		t.Fatalf("error = %+v; want METHOD_NOT_ALLOWED", env.Error)
	}
}

func TestRPCJSONHandler_unsupportedMediaType(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, web.TestRPCURLForTest("core.user", "search"), strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	web.ApplyDefaultRPCHeadersForTest(req)
	web.RPCJSONHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401 before media type check without auth", rr.Code)
	}
}

func TestRPCJSONHandler_emptyBody(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, web.TestRPCURLForTest("core.user", "search"), strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	web.ApplyDefaultRPCHeadersForTest(req)
	web.RPCJSONHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusUnauthorized)
	}
	decodeRPCEnvelope(t, rr)
}

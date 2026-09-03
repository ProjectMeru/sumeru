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
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  *api.RPCError   `json:"error"`
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
	req := httptest.NewRequest(http.MethodPost, "/api/rpc", strings.NewReader(`{"model":"core.user","method":"search","args":[]}`))
	req.Header.Set("Content-Type", "application/json")
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusUnauthorized)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.OK {
		t.Fatal("ok = true; want false")
	}
	if env.Error == nil || env.Error.Code != api.CodeUnauthorized {
		t.Fatalf("error = %+v; want UNAUTHORIZED", env.Error)
	}
}

func TestRPCJSONHandler_nonPostJSONEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/rpc", nil)
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.OK {
		t.Fatal("ok = true; want false")
	}
	if env.Error == nil || env.Error.Code != api.CodeMethodNotAllowed {
		t.Fatalf("error = %+v; want METHOD_NOT_ALLOWED", env.Error)
	}
}

func TestRPCJSONHandler_unsupportedMediaType(t *testing.T) {
	// Handler checks auth before Content-Type; unauthenticated requests still get 401 first.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rpc", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	web.RPCJSONHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401 before media type check without auth", rr.Code)
	}
}

func TestRPCJSONHandler_emptyBody(t *testing.T) {
	// Without auth, empty body still yields 401 (auth runs before dispatch).
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rpc", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	web.RPCJSONHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusUnauthorized)
	}
	decodeRPCEnvelope(t, rr)
}

package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/api"
	"sumeru/core/server/web"
)

func TestRPCJSONHandler_sessionRequiresCSRF(t *testing.T) {
	web.SetTestSessionUserIDForTest(42)
	t.Cleanup(web.ResetTestSessionUserIDForTest)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/rpc", strings.NewReader(`{"model":"core.user","method":"search","args":[]}`))
	req.Header.Set("Content-Type", "application/json")
	web.RPCJSONHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusForbidden)
	}
	env := decodeRPCEnvelope(t, rr)
	if env.OK {
		t.Fatal("ok = true; want false")
	}
	if env.Error == nil || env.Error.Code != api.CodeAccessDenied {
		t.Fatalf("error = %+v; want ACCESS_DENIED", env.Error)
	}
}

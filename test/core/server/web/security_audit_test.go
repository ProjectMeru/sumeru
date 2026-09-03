package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/web"
)

func TestCSRFTokenForRequest_emptyWithoutSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/web/home", nil)
	if tok := web.CSRFTokenForRequest(req); tok != "" {
		t.Fatalf("expected empty token without session, got %q", tok)
	}
}

func TestValidateCSRF_rejectsMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/web/company/switch", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()
	if web.ValidateCSRF(req) {
		t.Fatal("expected CSRF validation to fail without session/token")
	}
}

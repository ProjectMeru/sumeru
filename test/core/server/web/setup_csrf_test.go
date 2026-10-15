package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/web"
)

func TestValidateSetupCSRF_acceptsHeader(t *testing.T) {
	const token = "abc123setupcsrf"
	req := httptest.NewRequest(http.MethodPost, web.TestSetupInitRoute, nil)
	req.AddCookie(&http.Cookie{Name: web.TestSetupCSRFCookie, Value: token})
	req.Header.Set("X-CSRF-Token", token)
	if !web.ValidateSetupCSRFForTest(req, "") {
		t.Fatal("expected setup CSRF to pass with header")
	}
}

func TestValidateSetupCSRF_rejectsMissingCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, web.TestSetupInitRoute, nil)
	req.Header.Set("X-CSRF-Token", "orphan")
	if web.ValidateSetupCSRFForTest(req, "") {
		t.Fatal("expected setup CSRF to fail without cookie")
	}
}

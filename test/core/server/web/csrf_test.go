package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestValidateCSRF_acceptsMatchingFormToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/web/logout", nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: "test-session-sid"})
	token := web.CSRFTokenForRequest(req)
	if token == "" {
		t.Fatal("expected CSRF token for session cookie")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = map[string][]string{"csrf_token": {token}}
	if !web.ValidateCSRF(req) {
		t.Fatal("expected CSRF validation to pass with matching form token")
	}
}

func TestValidateCSRF_acceptsMatchingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/web/logout", nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: "header-session-sid"})
	token := web.CSRFTokenForRequest(req)
	req.Header.Set("X-CSRF-Token", token)
	if !web.ValidateCSRF(req) {
		t.Fatal("expected CSRF validation to pass with matching header")
	}
}

func TestValidateProductionCSRFSecret_devModeSkips(t *testing.T) {
	prevDev := config.AppConfig.DevMode
	prevSecret := config.AppConfig.CSRFSecret
	config.AppConfig.DevMode = true
	config.AppConfig.CSRFSecret = ""
	t.Cleanup(func() {
		config.AppConfig.DevMode = prevDev
		config.AppConfig.CSRFSecret = prevSecret
	})
	if err := web.ValidateProductionCSRFSecretForTest(); err != nil {
		t.Fatalf("dev mode should skip csrf_secret requirement: %v", err)
	}
}

func TestValidateProductionCSRFSecret_prodRequiresSecret(t *testing.T) {
	prevDev := config.AppConfig.DevMode
	prevSecret := config.AppConfig.CSRFSecret
	config.AppConfig.DevMode = false
	config.AppConfig.CSRFSecret = ""
	t.Cleanup(func() {
		config.AppConfig.DevMode = prevDev
		config.AppConfig.CSRFSecret = prevSecret
	})
	err := web.ValidateProductionCSRFSecretForTest()
	if err == nil || !strings.Contains(err.Error(), "csrf_secret") {
		t.Fatalf("prod without csrf_secret should fail, got %v", err)
	}
}

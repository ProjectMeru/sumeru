package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestSessionTouchPost_disabledWithoutConfig(t *testing.T) {
	prev := config.AppConfig.CloseSessionOnLastTab
	config.AppConfig.CloseSessionOnLastTab = false
	t.Cleanup(func() { config.AppConfig.CloseSessionOnLastTab = prev })

	req := httptest.NewRequest(http.MethodPost, web.TestSessionTouchRoute, nil)
	rec := httptest.NewRecorder()
	web.SessionTouchPostForTest(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d want 204", rec.Code)
	}
}

func TestSessionTouchPost_requiresCSRFWhenEnabled(t *testing.T) {
	prev := config.AppConfig.CloseSessionOnLastTab
	config.AppConfig.CloseSessionOnLastTab = true
	t.Cleanup(func() { config.AppConfig.CloseSessionOnLastTab = prev })

	req := httptest.NewRequest(http.MethodPost, web.TestSessionTouchRoute, strings.NewReader("csrf_token=bad"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	web.SessionTouchPostForTest(rec, req)
	if rec.Code != http.StatusFound && rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		if rec.Code == http.StatusNoContent {
			t.Fatal("expected auth/CSRF failure without session")
		}
	}
}

func TestBrowserSeenExpiredForTest(t *testing.T) {
	prev := config.AppConfig.CloseSessionOnLastTab
	config.AppConfig.CloseSessionOnLastTab = true
	config.AppConfig.SessionBrowserGraceSeconds = 60
	t.Cleanup(func() {
		config.AppConfig.CloseSessionOnLastTab = prev
		config.AppConfig.SessionBrowserGraceSeconds = 0
	})

	if web.BrowserSeenExpiredForTest(false, 0) {
		t.Fatal("null seen should not expire")
	}
}

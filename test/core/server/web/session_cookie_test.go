package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestBuildSessionCookie_hostPrefixWhenForcedSecure(t *testing.T) {
	prev := config.AppConfig.ForceSecureCookies
	config.AppConfig.ForceSecureCookies = true
	t.Cleanup(func() { config.AppConfig.ForceSecureCookies = prev })

	cookie := web.BuildSessionCookieForTest("abc", false)
	if cookie.Name != "__Host-sumeru_session" {
		t.Fatalf("cookie name = %q, want __Host-sumeru_session", cookie.Name)
	}
	if !cookie.Secure {
		t.Fatal("expected Secure=true with force_secure_cookies")
	}
}

func TestSessionCookieFromRequest_prefersHostCookie(t *testing.T) {
	prev := config.AppConfig.ForceSecureCookies
	config.AppConfig.ForceSecureCookies = true
	t.Cleanup(func() {
		config.AppConfig.ForceSecureCookies = prev
	})

	req := httptest.NewRequest(http.MethodGet, "/web/home", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-sumeru_session", Value: "sid-host"})
	req.AddCookie(&http.Cookie{Name: "sumeru_session", Value: "sid-legacy"})
	sid, name := web.SessionCookieFromRequestForTest(req)
	if sid != "sid-host" || name != "__Host-sumeru_session" {
		t.Fatalf("sid=%q name=%q", sid, name)
	}
}

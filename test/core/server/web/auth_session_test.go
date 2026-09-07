package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/web"
)

func TestBuildSessionCookieSessionScope(t *testing.T) {
	cookie := web.BuildSessionCookieForTest("session-id", false)
	if cookie.Name != web.TestSessionCookieName {
		t.Fatalf("cookie name = %q want %q", cookie.Name, web.TestSessionCookieName)
	}
	if cookie.MaxAge != 0 {
		t.Fatalf("session cookie MaxAge = %d, want 0 (browser session scope)", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Fatal("session cookie should be HttpOnly")
	}
}

func TestBuildSessionCookieDelete(t *testing.T) {
	cookie := web.BuildSessionCookieForTest("", true)
	if cookie.MaxAge != -1 {
		t.Fatalf("delete cookie MaxAge = %d, want -1", cookie.MaxAge)
	}
}

func TestAuthViaSessionUsesTestOverride(t *testing.T) {
	web.SetTestSessionUserIDForTest(7)
	t.Cleanup(web.ResetTestSessionUserIDForTest)

	req := httptest.NewRequest(http.MethodGet, web.TestHomeRoute, nil)
	if !web.AuthViaSessionForTest(req) {
		t.Fatal("AuthViaSession should be true when test override is set")
	}
	if got := web.SessionUserIDForTest(req); got != 7 {
		t.Fatalf("SessionUserID = %d, want 7", got)
	}
}

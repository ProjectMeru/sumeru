package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestSafePathNext(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		fallback string
		want     string
	}{
		{name: "valid path", rawURL: "/web/home", fallback: "/fallback", want: "/web/home"},
		{name: "empty uses fallback", rawURL: "  ", fallback: "/fallback", want: "/fallback"},
		{name: "protocol-relative rejected", rawURL: "//evil.com", fallback: "/fallback", want: "/fallback"},
		{name: "absolute URL rejected", rawURL: "https://evil.com", fallback: "/fallback", want: "/fallback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := web.SafePathNext(tt.rawURL, tt.fallback); got != tt.want {
				t.Fatalf("web.SafePathNext(%q) = %q, want %q", tt.rawURL, got, tt.want)
			}
		})
	}
}

func TestSafeWebNext(t *testing.T) {
	if got := web.SafeWebNext("/web?id=1", web.TestHomeRoute); got != "/web?id=1" {
		t.Fatalf("expected /web path, got %q", got)
	}
	if got := web.SafeWebNext("/login", web.TestHomeRoute); got != web.TestHomeRoute {
		t.Fatalf("non-/web path should fall back to home, got %q", got)
	}
}

func TestUrlWithQueryParam(t *testing.T) {
	got, err := web.URLWithQueryParam("/web/home", web.TestFlashMessageParam, "saved")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/web/home?msg=saved" {
		t.Fatalf("unexpected URL: %q", got)
	}

	got, err = web.URLWithQueryParam("/web/home?menu_id=3", web.TestFlashMessageParam, "saved")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "menu_id=3") || !strings.Contains(got, "msg=saved") {
		t.Fatalf("expected existing query preserved with msg, got %q", got)
	}
}

func TestRedirectWithWebMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", web.TestResetPasswordRoute, nil)
	web.RedirectWithWebMessage(recorder, request, "/web/settings", web.TestResetPasswordMsg)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	location := recorder.Header().Get("Location")
	if location != "/web/settings?msg="+web.TestResetPasswordMsg {
		t.Fatalf("unexpected Location: %q", location)
	}
}

func TestRequirePOST(t *testing.T) {
	recorder := httptest.NewRecorder()
	getRequest := httptest.NewRequest("GET", "/web", nil)
	if web.RequirePOST(recorder, getRequest) {
		t.Fatal("GET should be rejected")
	}
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	postRequest := httptest.NewRequest("POST", "/web", nil)
	if !web.RequirePOST(recorder, postRequest) {
		t.Fatal("POST should be accepted")
	}
}

func TestParsePostForm(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/web", strings.NewReader("%"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if web.ParsePostForm(recorder, request) {
		t.Fatal("malformed form should be rejected")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

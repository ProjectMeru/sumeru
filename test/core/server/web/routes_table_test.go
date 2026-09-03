package web_test

import (
	"net/http"
	"net/http/httptest"
	"sumeru/core/server/web"
	"testing"
)

func TestServeMuxOrDefault(t *testing.T) {
	customMux := http.NewServeMux()
	if got := web.ServeMuxOrDefault(customMux); got != customMux {
		t.Fatal("expected provided mux to be returned")
	}
	if got := web.ServeMuxOrDefault(nil); got != http.DefaultServeMux {
		t.Fatal("expected default serve mux when nil")
	}
}

func TestRootRedirectHome(t *testing.T) {
	recorder := httptest.NewRecorder()
	web.RootRedirectHome(recorder, httptest.NewRequest(http.MethodGet, web.TestRootRoute, nil))
	if recorder.Code != http.StatusFound {
		t.Fatalf("status=%d want %d", recorder.Code, http.StatusFound)
	}
	if location := recorder.Header().Get("Location"); location != web.TestHomeRoute {
		t.Fatalf("location=%q want %q", location, web.TestHomeRoute)
	}
}

func TestRootRedirectSetup(t *testing.T) {
	recorder := httptest.NewRecorder()
	web.RootRedirectSetup(recorder, httptest.NewRequest(http.MethodGet, web.TestRootRoute, nil))
	if location := recorder.Header().Get("Location"); location != web.TestSetupRoute {
		t.Fatalf("location=%q want %q", location, web.TestSetupRoute)
	}
}

func TestRedirectFoundTo(t *testing.T) {
	recorder := httptest.NewRecorder()
	web.RedirectFoundTo(web.TestAppsRoute)(recorder, httptest.NewRequest(http.MethodGet, web.TestAppsRoute+"/", nil))
	if location := recorder.Header().Get("Location"); location != web.TestAppsRoute {
		t.Fatalf("location=%q want %q", location, web.TestAppsRoute)
	}
}

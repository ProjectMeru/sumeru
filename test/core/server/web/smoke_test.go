package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sumeru/core/server/web"
)

func TestSmoke_unauthenticatedWebRedirectsToLogin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/web", web.WebHandler)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web: status %d; want %d", rr.Code, http.StatusFound)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil || !strings.HasPrefix(u.Path, "/web/login") {
		t.Fatalf("Location %q: want path prefix /web/login", loc)
	}
	if u.Query().Get("next") == "" {
		t.Fatalf("Location %q: want non-empty next query", loc)
	}
}

func TestSmoke_unauthenticatedSwcWorkspaceRedirectsToLogin(t *testing.T) {
	mux := http.NewServeMux()
	web.RegisterAppRoutes(mux)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/swc/workspace?action=1", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web/swc/workspace: status %d; want %d", rr.Code, http.StatusFound)
	}
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "/web/login") {
		t.Fatalf("Location %q: want /web/login", loc)
	}
}

func TestSmoke_registerAppRoutesRootRedirectsToHome(t *testing.T) {
	mux := http.NewServeMux()
	web.RegisterAppRoutes(mux)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /: status %d; want %d", rr.Code, http.StatusFound)
	}
	if g, w := rr.Header().Get("Location"), "/web/home"; g != w {
		t.Fatalf("Location %q; want %q", g, w)
	}
}

func TestSmoke_registerAppRoutesWebAppsTrailingSlashRedirect(t *testing.T) {
	mux := http.NewServeMux()
	web.RegisterAppRoutes(mux)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/apps/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web/apps/: status %d; want %d", rr.Code, http.StatusFound)
	}
	if g, w := rr.Header().Get("Location"), "/web/apps"; g != w {
		t.Fatalf("Location %q; want %q", g, w)
	}
}

func TestSmoke_settingsAppLogsNotCaughtByPrefixRedirect(t *testing.T) {
	mux := http.NewServeMux()
	web.RegisterAppRoutes(mux)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/settings/app-logs", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web/settings/app-logs: status %d; want %d", rr.Code, http.StatusFound)
	}
	loc := rr.Header().Get("Location")
	if strings.Contains(loc, "/web/settings") && !strings.Contains(loc, "/web/settings/app-logs") && !strings.Contains(loc, "/web/login") {
		t.Fatalf("Location %q: must not redirect to settings hub (/web/settings)", loc)
	}
	if !strings.Contains(loc, "/web/login") {
		t.Fatalf("Location %q: unauthenticated request should redirect to login", loc)
	}
}

func TestSmoke_settingsLegacyAliasRedirectsToHub(t *testing.T) {
	mux := http.NewServeMux()
	web.RegisterAppRoutes(mux)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/settings/home", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web/settings/home: status %d; want %d", rr.Code, http.StatusFound)
	}
	if g, w := rr.Header().Get("Location"), "/web/settings"; g != w {
		t.Fatalf("Location %q; want %q", g, w)
	}
}

func TestSmoke_apiHealth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", web.APIHealthHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET /api/health: status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "true") || !strings.Contains(body, "ok") {
		t.Fatalf("body: %s", body)
	}
}

//go:build integration

package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/web"
)

func rootMenuIDForModule(t *testing.T, moduleName string) int {
	t.Helper()
	table := orm.MustQuotedTableName("sys.menu")
	var id int
	err := orm.DB.QueryRow(
		`SELECT id FROM `+table+` WHERE module = $1 AND parent_id IS NULL ORDER BY sequence ASC, id ASC LIMIT 1`,
		moduleName,
	).Scan(&id)
	if err != nil {
		return 0
	}
	return id
}

func TestModuleEntry_crmRedirectsToWorkspace(t *testing.T) {
	integrationDB(t)
	userID := existingActiveUserID(t)
	sid := "test-module-entry-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { _ = web.DeleteTestSessionForTest(sid) })
	if err := web.InsertTestSessionForTest(sid, userID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	mux := http.NewServeMux()
	web.RegisterAppRoutes(mux)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/crm", nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: sid})
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("GET /web/crm: status %d; want %d body=%s", rr.Code, http.StatusFound, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("Location parse: %v", err)
	}
	if u.Path != web.TestWorkspaceRoute {
		t.Fatalf("Location path %q; want %q", u.Path, web.TestWorkspaceRoute)
	}
	q := u.Query()
	if q.Get(web.TestWorkspaceActionParam) == "" {
		t.Fatalf("Location %q: missing action param", loc)
	}
	if q.Get(web.TestWorkspaceMenuIDParam) == "" {
		t.Fatalf("Location %q: missing menu_id param", loc)
	}
}

func TestWorkspace_menuIDOnlyUsesAccessibleAction(t *testing.T) {
	integrationDB(t)
	userID := existingActiveUserID(t)
	sid := "test-menu-folder-acl-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { _ = web.DeleteTestSessionForTest(sid) })
	if err := web.InsertTestSessionForTest(sid, userID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	rootMenuID := rootMenuIDForModule(t, "crm")
	if rootMenuID <= 0 {
		t.Skip("crm root menu not installed")
	}

	mux := http.NewServeMux()
	web.RegisterAppRoutes(mux)
	h := web.SecurityMiddleware(mux)

	rr := httptest.NewRecorder()
	reqURL := web.TestWorkspaceRoute + "?" + web.TestWorkspaceMenuIDParam + "=" + strconv.Itoa(rootMenuID)
	req := httptest.NewRequest(http.MethodGet, reqURL, nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: sid})
	h.ServeHTTP(rr, req)

	if rr.Code == http.StatusFound && strings.HasPrefix(rr.Header().Get("Location"), web.TestHomeRoute) {
		t.Fatalf("menu_id-only workspace for crm root redirected home; admin should resolve an accessible action")
	}
	if rr.Code != http.StatusOK && rr.Code != http.StatusFound {
		t.Fatalf("unexpected status %d body=%s", rr.Code, rr.Body.String())
	}
}

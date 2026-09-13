//go:build integration

package web_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"sumeru/core/orm"
	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func integrationDB(t *testing.T) {
	t.Helper()
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	preflight, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := preflight.Ping(); err != nil {
		_ = preflight.Close()
		t.Fatalf("db ping: %v", err)
	}
	_ = preflight.Close()

	orm.InitDBWithPool(dsn, orm.DBPoolSettings{MaxOpenConns: 5, MaxIdleConns: 2})
	if !orm.IsInitialized() {
		t.Skip("database schema not bootstrapped (run sumeru -i base)")
	}
}

func existingActiveUserID(t *testing.T) int {
	t.Helper()
	userTable := orm.MustQuotedTableName("core.user")
	var userID int
	err := orm.DB.QueryRow(
		`SELECT id FROM `+userTable+` WHERE active = true ORDER BY id LIMIT 1`,
	).Scan(&userID)
	if err != nil || userID <= 0 {
		t.Fatalf("need at least one active core.user: %v", err)
	}
	return userID
}

func insertTestUser(t *testing.T) int {
	t.Helper()
	userTable := orm.MustQuotedTableName("core.user")
	login := fmt.Sprintf("web_sess_test_%d", time.Now().UnixNano())
	var userID int
	err := orm.DB.QueryRow(
		`INSERT INTO `+userTable+` (login, name, active, password, user_type)
		 VALUES ($1, $2, true, '', 'internal') RETURNING id`,
		login, "Web Session Test",
	).Scan(&userID)
	if err != nil || userID <= 0 {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = orm.DB.Exec(`DELETE FROM `+userTable+` WHERE id = $1`, userID)
	})
	return userID
}

func TestResolveSessionActiveUser(t *testing.T) {
	integrationDB(t)
	userID := existingActiveUserID(t)
	sid := "test-active-session-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { _ = web.DeleteTestSessionForTest(sid) })

	if err := web.InsertTestSessionForTest(sid, userID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, web.TestHomeRoute, nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: sid})
	gotUID, clearCookie := web.ResolveSessionFromCookieForTest(req)
	if clearCookie {
		t.Fatal("expected clearCookie false for active user session")
	}
	if gotUID != userID {
		t.Fatalf("ResolveSessionFromCookie uid = %d, want %d", gotUID, userID)
	}
}

func TestSecurityMiddlewareClearsCookieForInvalidSession(t *testing.T) {
	integrationDB(t)

	req := httptest.NewRequest(http.MethodGet, web.TestHomeRoute, nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: "invalid-or-expired-sid"})

	rec := httptest.NewRecorder()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid := web.SessionUserIDForTest(r); uid != 0 {
			t.Fatalf("expected uid 0 for invalid session, got %d", uid)
		}
		w.WriteHeader(http.StatusOK)
	})
	web.SecurityMiddleware(inner).ServeHTTP(rec, req)

	setCookies := rec.Result().Header.Values("Set-Cookie")
	foundClear := false
	for _, raw := range setCookies {
		if strings.Contains(raw, web.TestSessionCookieName+"=") {
			foundClear = true
			break
		}
	}
	if !foundClear {
		t.Fatalf("expected session cookie clear in Set-Cookie, got %v", setCookies)
	}
}

func TestLoginPost_rejectsBadCredentials(t *testing.T) {
	integrationDB(t)
	root := moduleRootFromIntegration(t)
	prevTemplates := config.AppConfig.TemplatesPath
	prevDev := config.AppConfig.DevMode
	config.AppConfig.TemplatesPath = filepath.Join(root, "core", "engine", "templates")
	config.AppConfig.DevMode = true
	t.Cleanup(func() {
		config.AppConfig.TemplatesPath = prevTemplates
		config.AppConfig.DevMode = prevDev
	})

	login := fmt.Sprintf("login_post_test_%d", time.Now().UnixNano())
	userTable := orm.MustQuotedTableName("core.user")
	var userID int
	err := orm.DB.QueryRow(
		`INSERT INTO `+userTable+` (login, name, active, password, user_type)
		 VALUES ($1, $2, true, '', 'internal') RETURNING id`,
		login, "Login Post Test",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() { _, _ = orm.DB.Exec(`DELETE FROM `+userTable+` WHERE id = $1`, userID) })

	web.ResetLoginLockoutForTest()
	t.Cleanup(web.ResetLoginLockoutForTest)

	csrfReq := httptest.NewRequest(http.MethodGet, web.TestLoginRoute, nil)
	csrfRec := httptest.NewRecorder()
	web.LoginGetForTest(csrfRec, csrfReq)
	csrfToken := loginCSRFCookieValue(csrfRec)
	if csrfToken == "" {
		csrfToken = hiddenInputValue(csrfRec.Body.String(), "csrf_token")
	}
	if csrfToken == "" {
		t.Fatal("expected login CSRF token from GET")
	}

	body := fmt.Sprintf("login=%s&password=wrong&next=%%2Fweb%%2Fhome&csrf_token=%s", login, csrfToken)
	req := httptest.NewRequest(http.MethodPost, web.TestLoginRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range csrfRec.Result().Cookies() {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	web.LoginPostForTest(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want 401 for bad credentials", rec.Code)
	}
}

func moduleRootFromIntegration(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestLogoutGetDestroysSession(t *testing.T) {
	integrationDB(t)
	userID := existingActiveUserID(t)
	sid := "test-logout-get-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { _ = web.DeleteTestSessionForTest(sid) })

	if err := web.InsertTestSessionForTest(sid, userID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, web.TestLogoutRoute, nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: sid})
	rec := httptest.NewRecorder()
	web.LogoutGetForTest(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status=%d want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != web.TestLoginRoute {
		t.Fatalf("Location=%q want %q", loc, web.TestLoginRoute)
	}
	count, err := web.CountTestSessionsForUserForTest(userID)
	if err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("session row count=%d want 0 after logout", count)
	}
	setCookies := rec.Result().Header.Values("Set-Cookie")
	foundClear := false
	for _, raw := range setCookies {
		if strings.Contains(raw, web.TestSessionCookieName+"=") {
			foundClear = true
			break
		}
	}
	if !foundClear {
		t.Fatalf("expected cleared session cookie, got %v", setCookies)
	}
}

func TestLogoutPostDestroysSessionWithCSRF(t *testing.T) {
	integrationDB(t)
	userID := existingActiveUserID(t)
	sid := "test-logout-post-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { _ = web.DeleteTestSessionForTest(sid) })

	if err := web.InsertTestSessionForTest(sid, userID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, web.TestLogoutRoute, strings.NewReader("csrf_token=placeholder"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: sid})
	csrf := web.CSRFTokenForRequestForTest(req)
	if csrf == "" {
		t.Fatal("expected non-empty CSRF token for session")
	}
	req = httptest.NewRequest(http.MethodPost, web.TestLogoutRoute, strings.NewReader("csrf_token="+csrf))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: sid})

	rec := httptest.NewRecorder()
	web.LogoutPostForTest(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d want 303", rec.Code)
	}
	count, err := web.CountTestSessionsForUserForTest(userID)
	if err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("session row count=%d want 0 after logout POST", count)
	}
}

func TestInactiveUserSessionRevoked(t *testing.T) {
	integrationDB(t)
	userID := insertTestUser(t)
	sid := "test-inactive-session-" + time.Now().Format("150405.000000")
	t.Cleanup(func() { _ = web.DeleteTestSessionForTest(sid) })

	if err := web.InsertTestSessionForTest(sid, userID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("insert session: %v", err)
	}

	userTable := orm.MustQuotedTableName("core.user")
	if _, err := orm.DB.Exec(`UPDATE `+userTable+` SET active = false WHERE id = $1`, userID); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, web.TestAPIRPCRoute, nil)
	req.AddCookie(&http.Cookie{Name: web.TestSessionCookieName, Value: sid})
	rec := httptest.NewRecorder()
	web.SecurityMiddleware(http.HandlerFunc(web.RPCJSONHandlerForTest)).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("RPC status = %d, want 401", rec.Code)
	}
	count, err := web.CountTestSessionsForUserForTest(userID)
	if err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("session row count = %d, want 0 after inactive revocation", count)
	}
	setCookies := rec.Result().Header.Values("Set-Cookie")
	foundClear := false
	for _, raw := range setCookies {
		if strings.Contains(raw, web.TestSessionCookieName+"=") {
			foundClear = true
			break
		}
	}
	if !foundClear {
		t.Fatalf("expected cleared session cookie, got %v", setCookies)
	}
}

package web_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestLoginGet_embedsLoginCSRFTokenInForm(t *testing.T) {
	root := sumeruModuleRoot(t)
	prevTemplates := config.AppConfig.TemplatesPath
	prevDev := config.AppConfig.DevMode
	config.AppConfig.TemplatesPath = filepath.Join(root, "core", "engine", "templates")
	config.AppConfig.DevMode = true
	t.Cleanup(func() {
		config.AppConfig.TemplatesPath = prevTemplates
		config.AppConfig.DevMode = prevDev
	})

	req := httptest.NewRequest(http.MethodGet, web.TestLoginRoute, nil)
	rec := httptest.NewRecorder()
	web.LoginGetForTest(rec, req)

	cookieToken := loginCSRFCookieValue(rec)
	formToken := hiddenInputValue(rec.Body.String(), "csrf_token")
	if cookieToken == "" {
		t.Fatal("expected Set-Cookie login CSRF token")
	}
	if formToken == "" {
		t.Fatal("expected csrf_token hidden field in login form")
	}
	if cookieToken != formToken {
		t.Fatalf("cookie token %q != form token %q", cookieToken, formToken)
	}
}

func TestRedirectToLogin_usesCleanURLAndNextCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/web/home", nil)
	rec := httptest.NewRecorder()
	web.RedirectToLoginForTest(rec, req, "/web/home?menu_id=2")

	if rec.Code != http.StatusFound {
		t.Fatalf("status=%d want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != web.TestLoginRoute {
		t.Fatalf("Location=%q want %q", loc, web.TestLoginRoute)
	}
	nextCookie := ""
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sumeru_login_next" {
			nextCookie = c.Value
		}
	}
	if nextCookie != "/web/home?menu_id=2" {
		t.Fatalf("login next cookie=%q", nextCookie)
	}
}

func TestLoginGet_stripsQueryNextToCleanURL(t *testing.T) {
	root := sumeruModuleRoot(t)
	prevTemplates := config.AppConfig.TemplatesPath
	prevDev := config.AppConfig.DevMode
	config.AppConfig.TemplatesPath = filepath.Join(root, "core", "engine", "templates")
	config.AppConfig.DevMode = true
	t.Cleanup(func() {
		config.AppConfig.TemplatesPath = prevTemplates
		config.AppConfig.DevMode = prevDev
	})

	req := httptest.NewRequest(http.MethodGet, web.TestLoginRoute+"?next=%2Fweb%2Fapps", nil)
	rec := httptest.NewRecorder()
	web.LoginGetForTest(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status=%d want redirect", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != web.TestLoginRoute {
		t.Fatalf("Location=%q want clean login URL", loc)
	}
}

func TestLoginGet_embedsNextFromCookie(t *testing.T) {
	root := sumeruModuleRoot(t)
	prevTemplates := config.AppConfig.TemplatesPath
	prevDev := config.AppConfig.DevMode
	config.AppConfig.TemplatesPath = filepath.Join(root, "core", "engine", "templates")
	config.AppConfig.DevMode = true
	t.Cleanup(func() {
		config.AppConfig.TemplatesPath = prevTemplates
		config.AppConfig.DevMode = prevDev
	})

	req := httptest.NewRequest(http.MethodGet, web.TestLoginRoute, nil)
	req.AddCookie(&http.Cookie{Name: "sumeru_login_next", Value: "/web/apps"})
	rec := httptest.NewRecorder()
	web.LoginGetForTest(rec, req)

	formNext := hiddenInputValue(rec.Body.String(), "next")
	if formNext != "/web/apps" {
		t.Fatalf("form next=%q want /web/apps", formNext)
	}
}

func TestValidateLoginCSRFForTest_acceptsMatchingCookieAndForm(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, web.TestLoginRoute, strings.NewReader("csrf_token=abc&login=u&password=p"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()
	req.AddCookie(&http.Cookie{Name: web.TestLoginCSRFCookie, Value: "abc"})
	if !web.ValidateLoginCSRFForTest(req) {
		t.Fatal("expected matching login CSRF token to validate")
	}
}

func loginCSRFCookieValue(rec *httptest.ResponseRecorder) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == web.TestLoginCSRFCookie && c.Value != "" {
			return c.Value
		}
	}
	return ""
}

func hiddenInputValue(html, name string) string {
	re := regexp.MustCompile(`name="` + regexp.QuoteMeta(name) + `" value="([^"]*)"`)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func sumeruModuleRoot(t *testing.T) string {
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

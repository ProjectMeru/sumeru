package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/web"
)

func TestParseContentPathAttachmentID(t *testing.T) {
	_, id, ok := web.ParseContentPathForTest("42")
	if !ok || id != 42 {
		t.Fatalf("parse attachment id: ok=%v id=%d", ok, id)
	}
}

func TestParseContentPathModelField(t *testing.T) {
	field, id, ok := web.ParseContentPathForTest("core.partner/name/7")
	if !ok || field != "name" || id != 7 {
		t.Fatalf("parse model field: ok=%v field=%q id=%d", ok, field, id)
	}
}

func TestContentHandlerModelFieldPath(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodGet, "/web/content/core.partner/name/1?download=0", nil)
	rec := httptest.NewRecorder()
	web.ContentHandlerForTest(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("unexpected success without db")
	}
}

func TestContentHandlerNotFound(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodGet, "/web/content/99999", nil)
	rec := httptest.NewRecorder()
	web.ContentHandlerForTest(rec, req)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestSafeContentDispositionInline(t *testing.T) {
	d := web.SafeContentDispositionInlineForTest("preview.pdf")
	if d == "" || !strings.Contains(d, "inline") {
		t.Fatalf("disposition=%q", d)
	}
}

func TestContentInlineAllowedPDF(t *testing.T) {
	if !web.ContentInlineAllowedForTest("application/pdf") {
		t.Fatal("pdf should allow inline")
	}
	if web.ContentInlineAllowedForTest("text/html") {
		t.Fatal("html must not inline")
	}
	if web.ContentInlineAllowedForTest("image/svg+xml") {
		t.Fatal("svg must not inline")
	}
	if !web.ContentInlineAllowedForTest("image/png") {
		t.Fatal("png should allow inline")
	}
}

package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/web"
)

func TestExportXLSXHandlerRequiresLogin(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/export/xlsx?model=core.user&fields=name", nil)
	web.ExportXLSXHandlerForTest(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
}

func TestResolveExportRequestMissingModel(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/export/xlsx?fields=name", nil)
	_, ok := web.ResolveExportRequestForTest(rr, req)
	if ok {
		t.Fatal("expected resolve export request to fail without model")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestExportTemplatePDFHandlerRequiresLogin(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/report/template-pdf?title=Demo", nil)
	web.ExportTemplatePDFHandlerForTest(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
}

func TestExportTemplatePDFHandlerMissingTitle(t *testing.T) {
	web.SetTestSessionUserIDForTest(1)
	defer web.ResetTestSessionUserIDForTest()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/report/template-pdf", nil)
	web.ExportTemplatePDFHandlerForTest(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestExportTemplatePDFHandlerReturnsPDF(t *testing.T) {
	web.SetTestSessionUserIDForTest(1)
	defer web.ResetTestSessionUserIDForTest()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/web/report/template-pdf?title=Invoice&subtitle=Acme", nil)
	web.ExportTemplatePDFHandlerForTest(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Fatalf("content-type = %q", ct)
	}
	body := rr.Body.Bytes()
	if len(body) < 100 || string(body[:4]) != "%PDF" {
		t.Fatalf("expected PDF bytes, got len=%d", len(body))
	}
}

package web_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/web"
)

func TestBulkBatchOwnedBy_acceptsMatch(t *testing.T) {
	batch := map[string]interface{}{"user_id": int64(7)}
	if !web.BulkBatchOwnedByForTest(batch, 7) {
		t.Fatal("expected owner match")
	}
}

func TestBulkBatchOwnedBy_rejectsWrongUser(t *testing.T) {
	batch := map[string]interface{}{"user_id": int64(1)}
	if web.BulkBatchOwnedByForTest(batch, 2) {
		t.Fatal("expected foreign owner to fail")
	}
	if web.BulkBatchOwnedByForTest(batch, 0) {
		t.Fatal("expected uid 0 to fail")
	}
}

func TestBulkBatchOwnedBy_rejectsMissingOwner(t *testing.T) {
	if web.BulkBatchOwnedByForTest(map[string]interface{}{}, 1) {
		t.Fatal("expected missing user_id to fail")
	}
}

func TestSwcImportHandlerMissingBatchRecord(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodGet, "/web/swc/import?batch=1", nil)
	rec := httptest.NewRecorder()
	web.SwcImportHandlerForTest(rec, req)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestSwcImportHandlerRequiresBatch(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodGet, "/web/swc/import", nil)
	rec := httptest.NewRecorder()
	web.SwcImportHandlerForTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestImportWizardPageRequiresBatch(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodGet, "/web/import", nil)
	rec := httptest.NewRecorder()
	web.ImportWizardPageHandlerForTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestBulkStatusHandlerRequiresID(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodGet, "/web/bulk/status", nil)
	rec := httptest.NewRecorder()
	web.BulkStatusHandlerForTest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestExportAsyncHandlerRequiresModel(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodPost, "/web/export/async", bytes.NewReader([]byte("csrf_token=x")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	web.ExportAsyncHandlerForTest(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected failure without model")
	}
}

func TestBulkDryRunHandlerRejectsBadBody(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodPost, "/web/bulk/dry-run", bytes.NewReader([]byte(`{"batch_id":0}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	web.BulkDryRunHandlerForTest(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected failure")
	}
}

func TestBulkSaveImportTemplateHandlerRejectsBadBody(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodPost, "/web/bulk/save-template", bytes.NewReader([]byte(`{"batch_id":1}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	web.BulkSaveImportTemplateHandlerForTest(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected failure without name")
	}
}

func TestBulkPreviewHandlerRejectsBadBody(t *testing.T) {
	web.ResetTestSessionUserIDForTest()
	t.Cleanup(web.ResetTestSessionUserIDForTest)
	web.SetTestSessionUserIDForTest(1)
	req := httptest.NewRequest(http.MethodPost, "/web/bulk/preview", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	web.BulkPreviewHandlerForTest(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
}

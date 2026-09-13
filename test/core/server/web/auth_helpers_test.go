package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/orm"
	"sumeru/core/server/web"
)

func TestRequireSystemAdmin_deniesNonAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/web/settings", nil)
	req = req.WithContext(orm.ContextWithUID(context.Background(), 2))
	rec := httptest.NewRecorder()

	if web.RequireSystemAdminForTest(rec, req, false) {
		t.Fatal("expected non-admin to be denied")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d want 403", rec.Code)
	}
}

func TestRequireModelAccess_deniesWithoutPermission(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/web/home", nil)
	req = req.WithContext(orm.ContextWithUID(context.Background(), 2))
	rec := httptest.NewRecorder()

	if web.RequireModelAccessForTest(rec, req, "app.log", "read") {
		t.Fatal("expected model access denial without DB permissions")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d want 403", rec.Code)
	}
}

package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/orm"
	"sumeru/core/server/web"
)

func TestRequireModuleManagerOrSystem_deniesNonManager(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/web/apps", nil)
	req = req.WithContext(orm.ContextWithUID(context.Background(), 2))
	rec := httptest.NewRecorder()

	if web.RequireModuleManagerOrSystemForTest(rec, req, false) {
		t.Fatal("expected non-manager user to be denied")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d want 403", rec.Code)
	}
}

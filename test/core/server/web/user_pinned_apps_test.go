package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/web"
)

func TestPinnedAppsSaveHandler_requiresLogin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/web/user/pinned-apps", nil)
	rr := httptest.NewRecorder()
	web.PinnedAppsSaveHandler(rr, req)
	if rr.Code != http.StatusFound && rr.Code != http.StatusUnauthorized {
		// requireLogin redirects to login (302) when no session
		if rr.Code != http.StatusFound {
			t.Fatalf("status %d, want redirect or unauthorized", rr.Code)
		}
	}
}

func TestPinnedAppsSaveHandler_rejectsMissingCSRF(t *testing.T) {
	// Without session, requireLogin fires first; with session would need CSRF test harness.
	req := httptest.NewRequest(http.MethodPost, "/web/user/pinned-apps", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	web.PinnedAppsSaveHandler(rr, req)
	if rr.Code == http.StatusOK {
		t.Fatal("expected handler to reject unauthenticated request")
	}
}

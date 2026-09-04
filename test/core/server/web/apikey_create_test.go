package web_test

import (
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestApiKeyTargetUserID(t *testing.T) {
	t.Run("falls back to session when targeting other user without admin", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/web/action/create_api_key", strings.NewReader("user_id=42"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		// No session → session UID is 0; non-admin cannot mint for user 42.
		if got := web.APIKeyTargetUserID(req); got != 0 {
			t.Fatalf("got %d, want 0 without system group", got)
		}
	})

	t.Run("falls back to zero when user_id missing", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/web/action/create_api_key", nil)
		if got := web.APIKeyTargetUserID(req); got != 0 {
			t.Fatalf("got %d, want 0 without session", got)
		}
	})
}

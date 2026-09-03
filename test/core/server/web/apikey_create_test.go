package web_test

import (
	"net/http/httptest"
	"strings"
	"sumeru/core/server/web"
	"testing"
)

func TestApiKeyTargetUserID(t *testing.T) {
	t.Run("uses form user_id when positive", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/web/action/create_api_key", strings.NewReader("user_id=42"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := web.APIKeyTargetUserID(req); got != 42 {
			t.Fatalf("got %d, want 42", got)
		}
	})

	t.Run("falls back to zero when user_id missing", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/web/action/create_api_key", nil)
		if got := web.APIKeyTargetUserID(req); got != 0 {
			t.Fatalf("got %d, want 0 without session", got)
		}
	})
}

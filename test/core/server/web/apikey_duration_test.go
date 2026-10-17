package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sumeru/core/server/web"
)

func TestAPIKeyExpiresAtFromForm(t *testing.T) {
	t.Run("missing duration", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("name=k"))
		_ = req.ParseForm()
		_, msg, err := web.APIKeyExpiresAtFromFormForTest(req)
		if err != nil || msg != "api_key_duration_required" {
			t.Fatalf("msg=%q err=%v", msg, err)
		}
	})
	t.Run("invalid duration", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("duration_days=abc"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		_, msg, _ := web.APIKeyExpiresAtFromFormForTest(req)
		if msg != "api_key_duration_invalid" {
			t.Fatalf("msg=%q", msg)
		}
	})
	t.Run("31 days rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.PostForm = url.Values{"duration_days": {"31"}}
		_, msg, _ := web.APIKeyExpiresAtFromFormForTest(req)
		if msg != "api_key_duration_invalid" {
			t.Fatalf("msg=%q", msg)
		}
	})
}

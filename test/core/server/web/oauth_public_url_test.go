package web_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

func TestPublicBaseURLUsesConfiguredOrigin(t *testing.T) {
	prev := config.AppConfig.WebBaseURL
	config.AppConfig.WebBaseURL = "https://erp.example.com/"
	t.Cleanup(func() { config.AppConfig.WebBaseURL = prev })

	req := httptest.NewRequest("GET", "http://evil.example/oauth", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	got, err := web.PublicBaseURLForTest(req)
	if err != nil || got != "https://erp.example.com" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestPublicBaseURLRejectsUntrustedRemote(t *testing.T) {
	prev := config.AppConfig.WebBaseURL
	config.AppConfig.WebBaseURL = ""
	t.Cleanup(func() { config.AppConfig.WebBaseURL = prev })

	req := httptest.NewRequest("GET", "http://evil.example/oauth", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	_, err := web.PublicBaseURLForTest(req)
	if err == nil || !strings.Contains(err.Error(), "web_base_url") {
		t.Fatalf("err %v", err)
	}
}

func TestPublicBaseURLAllowsLoopback(t *testing.T) {
	prev := config.AppConfig.WebBaseURL
	config.AppConfig.WebBaseURL = ""
	t.Cleanup(func() { config.AppConfig.WebBaseURL = prev })

	req := httptest.NewRequest("GET", "http://127.0.0.1:8080/oauth", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	got, err := web.PublicBaseURLForTest(req)
	if err != nil || got != "http://127.0.0.1:8080" {
		t.Fatalf("got %q err %v", got, err)
	}
}

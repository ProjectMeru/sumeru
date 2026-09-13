package server_test

import (
	"testing"

	"sumeru/core/server"
)

func TestListenAddr(t *testing.T) {
	if got := server.ListenAddrForTest("", "8080"); got != ":8080" {
		t.Fatalf("expected :8080, got %q", got)
	}
	if got := server.ListenAddrForTest("127.0.0.1", "8080"); got != "127.0.0.1:8080" {
		t.Fatalf("expected 127.0.0.1:8080, got %q", got)
	}
}

func TestSetupListenAddr(t *testing.T) {
	cfg := server.ConfigForTest("", "8080", true)
	if got := server.SetupListenAddrForTest(cfg); got != "127.0.0.1:8080" {
		t.Fatalf("expected localhost bind, got %q", got)
	}
	cfg2 := server.ConfigForTest("0.0.0.0", "8080", true)
	if got := server.SetupListenAddrForTest(cfg2); got != "127.0.0.1:8080" {
		t.Fatalf("setup_localhost_only must force 127.0.0.1, got %q", got)
	}
	cfg3 := server.ConfigForTest("0.0.0.0", "8080", false)
	if got := server.SetupListenAddrForTest(cfg3); got != "0.0.0.0:8080" {
		t.Fatalf("expected explicit interface when not localhost-only, got %q", got)
	}
	cfg4 := server.ConfigForTest("", "8080", false)
	if got := server.SetupListenAddrForTest(cfg4); got != ":8080" {
		t.Fatalf("expected all-interfaces bind, got %q", got)
	}
}

func TestValidateSetupModeConfig(t *testing.T) {
	cfg := server.ConfigForSetupTest("", "8080", "", false)
	if err := server.ValidateSetupModeConfigForTest(cfg); err == nil {
		t.Fatal("expected error when setup_localhost_only=false and setup_token empty")
	}
	cfg.SetupToken = "secret"
	if err := server.ValidateSetupModeConfigForTest(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg2 := server.ConfigForSetupTest("", "8080", "", true)
	if err := server.ValidateSetupModeConfigForTest(cfg2); err != nil {
		t.Fatalf("empty token allowed when setup_localhost_only=true: %v", err)
	}
}

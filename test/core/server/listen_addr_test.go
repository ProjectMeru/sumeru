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
	if got := server.SetupListenAddrForTest(cfg2); got != "0.0.0.0:8080" {
		t.Fatalf("expected explicit interface, got %q", got)
	}
}

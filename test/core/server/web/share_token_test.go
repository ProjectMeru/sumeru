package web_test

import (
	"testing"
	"time"

	"sumeru/core/server/web"
)

func TestShareTokenRoundTrip(t *testing.T) {
	token, err := web.MintShareTokenForTest(5, "core.partner", 42)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := web.ParseShareTokenForTest(token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.IssuerUID != 5 || parsed.Model != "core.partner" || parsed.ResID != 42 {
		t.Fatalf("parsed: %+v", parsed)
	}
}

func TestShareTokenTamperRejected(t *testing.T) {
	token, err := web.MintShareTokenForTest(1, "core.partner", 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := web.ParseShareTokenForTest(token + "x"); err == nil {
		t.Fatal("expected tamper error")
	}
}

func TestShareTokenExpired(t *testing.T) {
	token, err := web.MintShareTokenForTestWithExpiry(1, "core.partner", 1, time.Now().UTC().Add(-time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := web.ParseShareTokenForTest(token); err == nil {
		t.Fatal("expected expiry error")
	}
}

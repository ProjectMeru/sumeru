package orm_test

import (
	"testing"
	"time"

	"sumeru/core/orm"
)

func TestValidateAPIKeyExpiresAt(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := orm.ValidateAPIKeyExpiresAt(time.Time{}, now); err == nil {
		t.Fatal("zero expiry should fail")
	}
	if err := orm.ValidateAPIKeyExpiresAt(now, now); err == nil {
		t.Fatal("past or equal expiry should fail")
	}
	if err := orm.ValidateAPIKeyExpiresAt(now.Add(31*24*time.Hour), now); err == nil {
		t.Fatal("over 30 days should fail")
	}
	if err := orm.ValidateAPIKeyExpiresAt(now.Add(7*24*time.Hour), now); err != nil {
		t.Fatalf("7 days should pass: %v", err)
	}
}

func TestExpiresAtFromDurationDays(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if _, err := orm.ExpiresAtFromDurationDays(0, now); err == nil {
		t.Fatal("0 days should fail")
	}
	if _, err := orm.ExpiresAtFromDurationDays(31, now); err == nil {
		t.Fatal("31 days should fail")
	}
	exp, err := orm.ExpiresAtFromDurationDays(30, now)
	if err != nil {
		t.Fatalf("30 days: %v", err)
	}
	if !exp.After(now) {
		t.Fatal("expected future expiry")
	}
}

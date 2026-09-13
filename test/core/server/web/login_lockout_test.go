package web_test

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"sumeru/core/server/web"
)

func TestLoginLockout_locksAfterMaxFailures(t *testing.T) {
	web.ResetLoginLockoutForTest()
	t.Cleanup(web.ResetLoginLockoutForTest)

	login := "lockout-test@example.com"
	for i := 0; i < 5; i++ {
		web.RecordLoginFailureForTest(login)
	}
	if !web.LoginLockedForTest(login) {
		t.Fatal("expected login locked after 5 failures")
	}
}

func TestLoginLockout_clearUnlocks(t *testing.T) {
	web.ResetLoginLockoutForTest()
	t.Cleanup(web.ResetLoginLockoutForTest)

	login := "unlock-test@example.com"
	for i := 0; i < 5; i++ {
		web.RecordLoginFailureForTest(login)
	}
	web.ClearLoginFailuresForTest(login)
	if web.LoginLockedForTest(login) {
		t.Fatal("expected login unlocked after clear")
	}
}

func TestComparePasswordConstantTime(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if web.ComparePasswordConstantTimeForTest(string(hash), "wrong") {
		t.Fatal("wrong password should not match")
	}
	if !web.ComparePasswordConstantTimeForTest(string(hash), "correct") {
		t.Fatal("correct password should match")
	}
	if web.ComparePasswordConstantTimeForTest("", "any") {
		t.Fatal("empty stored hash should not match")
	}
}

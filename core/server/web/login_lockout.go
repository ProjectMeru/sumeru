package web

import (
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ponytail: in-memory lockout map; single-process only; upgrade to DB/redis for multi-instance.
var (
	loginLockoutMu sync.Mutex
	loginFailures  = map[string][]time.Time{}
)

const (
	loginLockoutMaxFailures = 5
	loginLockoutWindow      = 15 * time.Minute
)

// bcrypt dummy hash (cost 10) for constant-time path on unknown login.
var dummyPasswordHash = mustDummyBcryptHash()

func mustDummyBcryptHash() string {
	h, err := bcrypt.GenerateFromPassword([]byte("sumeru-timing-dummy-password"), bcrypt.DefaultCost)
	if err != nil {
		panic("login lockout: dummy bcrypt: " + err.Error())
	}
	return string(h)
}

func normalizeLoginKey(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

func loginLocked(login string) bool {
	key := normalizeLoginKey(login)
	if key == "" {
		return false
	}
	loginLockoutMu.Lock()
	defer loginLockoutMu.Unlock()
	cutoff := time.Now().Add(-loginLockoutWindow)
	attempts := pruneAttempts(loginFailures[key], cutoff)
	loginFailures[key] = attempts
	return len(attempts) >= loginLockoutMaxFailures
}

func recordLoginFailure(login string) {
	key := normalizeLoginKey(login)
	if key == "" {
		return
	}
	loginLockoutMu.Lock()
	defer loginLockoutMu.Unlock()
	cutoff := time.Now().Add(-loginLockoutWindow)
	loginFailures[key] = append(pruneAttempts(loginFailures[key], cutoff), time.Now())
}

func clearLoginFailures(login string) {
	key := normalizeLoginKey(login)
	if key == "" {
		return
	}
	loginLockoutMu.Lock()
	delete(loginFailures, key)
	loginLockoutMu.Unlock()
}

func pruneAttempts(attempts []time.Time, cutoff time.Time) []time.Time {
	out := attempts[:0]
	for _, t := range attempts {
		if t.After(cutoff) {
			out = append(out, t)
		}
	}
	return out
}

func comparePasswordConstantTime(storedHash, plain string) bool {
	storedHash = strings.TrimSpace(storedHash)
	if storedHash == "" {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(plain))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plain)) == nil
}

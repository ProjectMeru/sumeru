package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
)

const (
	csrfHeaderName = "X-CSRF-Token"
	csrfFormField  = "csrf_token"
)

var (
	csrfSecretMu sync.RWMutex
	csrfSecret   []byte
)

func csrfKey() []byte {
	csrfSecretMu.RLock()
	if len(csrfSecret) > 0 {
		defer csrfSecretMu.RUnlock()
		return csrfSecret
	}
	csrfSecretMu.RUnlock()

	csrfSecretMu.Lock()
	defer csrfSecretMu.Unlock()
	if len(csrfSecret) == 0 {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			panic("csrf: crypto/rand failed: " + err.Error())
		}
		csrfSecret = b
	}
	return csrfSecret
}

func sessionIDFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}

// CSRFTokenForRequest returns the per-session CSRF token (empty when not logged in).
func CSRFTokenForRequest(r *http.Request) string {
	sessionID := sessionIDFromRequest(r)
	if sessionID == "" {
		return ""
	}
	mac := hmac.New(sha256.New, csrfKey())
	_, _ = mac.Write([]byte(sessionID))
	return hex.EncodeToString(mac.Sum(nil)[:16])
}

// ValidateCSRF checks the csrf_token form field or X-CSRF-Token header against the session-bound token.
func ValidateCSRF(r *http.Request) bool {
	expected := CSRFTokenForRequest(r)
	if expected == "" {
		return false
	}
	got := r.PostFormValue(csrfFormField)
	if got == "" {
		got = r.Header.Get(csrfHeaderName)
	}
	return got != "" && hmac.Equal([]byte(got), []byte(expected))
}

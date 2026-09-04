package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"

	"sumeru/core/server/config"
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
		if configured := strings.TrimSpace(config.AppConfig.CSRFSecret); configured != "" {
			csrfSecret = []byte(configured)
		} else {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				panic("csrf: crypto/rand failed: " + err.Error())
			}
			csrfSecret = b
		}
	}
	return csrfSecret
}

// InitCSRFSecret loads csrf_secret from config (call after LoadConfig). Empty keeps ephemeral key.
func InitCSRFSecret() {
	csrfSecretMu.Lock()
	defer csrfSecretMu.Unlock()
	if configured := strings.TrimSpace(config.AppConfig.CSRFSecret); configured != "" {
		csrfSecret = []byte(configured)
	}
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

// ValidateCSRF checks the csrf_token form field, query param, or X-CSRF-Token header against the session-bound token.
func ValidateCSRF(r *http.Request) bool {
	expected := CSRFTokenForRequest(r)
	if expected == "" {
		return false
	}
	got := r.PostFormValue(csrfFormField)
	if got == "" {
		got = r.Header.Get(csrfHeaderName)
	}
	if got == "" {
		got = strings.TrimSpace(r.URL.Query().Get(csrfFormField))
	}
	return got != "" && hmac.Equal([]byte(got), []byte(expected))
}

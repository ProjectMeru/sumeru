package web

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"sumeru/core/server/config"
)

const loginCSRFCookie = "sumeru_login_csrf"

func setLoginCSRFCookie(w http.ResponseWriter) {
	token := newLoginCSRFToken()
	http.SetCookie(w, &http.Cookie{
		Name:     loginCSRFCookie,
		Value:    token,
		Path:     loginRoute,
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   sessionCookieSecure(),
	})
}

func newLoginCSRFToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic("login csrf: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(buf)
}

func validateLoginCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(loginCSRFCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	got := strings.TrimSpace(r.PostFormValue(csrfFormField))
	if got == "" {
		return false
	}
	return hmac.Equal([]byte(got), []byte(cookie.Value))
}

func clearLoginCSRFCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     loginCSRFCookie,
		Value:    "",
		Path:     loginRoute,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   sessionCookieSecure(),
	})
}

func sessionCookieSecure() bool {
	if config.AppConfig.ForceSecureCookies {
		return true
	}
	return !config.AppConfig.DevMode
}

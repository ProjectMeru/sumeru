package web

import (
	"net/http"

	"sumeru/core/server/config"
)

func buildNamedCookie(name, value, path string, maxAge int, sameSite http.SameSite, httpOnly, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		SameSite: sameSite,
		Secure:   secure,
	}
}

func setNamedCookie(w http.ResponseWriter, name, value, path string, maxAge int, sameSite http.SameSite) {
	http.SetCookie(w, buildNamedCookie(name, value, path, maxAge, sameSite, true, sessionCookieSecure()))
}

func clearNamedCookie(w http.ResponseWriter, name, path string, sameSite http.SameSite) {
	http.SetCookie(w, buildNamedCookie(name, "", path, -1, sameSite, true, sessionCookieSecure()))
}

func sessionSameSite() http.SameSite {
	if config.AppConfig.SessionCookieStrictSameSite {
		return http.SameSiteStrictMode
	}
	return http.SameSiteLaxMode
}

func effectiveSessionCookieName() string {
	if config.AppConfig.ForceSecureCookies {
		return hostPrefixedSessionCookieName
	}
	return sessionCookieName
}

const hostPrefixedSessionCookieName = "__Host-sumeru_session"

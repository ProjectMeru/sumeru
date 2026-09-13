package web

import (
	"net/http"
	"strings"
)

const loginNextCookie = "sumeru_login_next"

func setLoginNextCookie(w http.ResponseWriter, returnTo string) {
	returnTo = SafePathNext(returnTo, homeRoute)
	setNamedCookie(w, loginNextCookie, returnTo, loginRoute, 600, http.SameSiteLaxMode)
}

func loginNextFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(loginNextCookie)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return SafePathNext(cookie.Value, homeRoute)
}

func resolveLoginNext(r *http.Request) string {
	if next := loginNextFromRequest(r); next != "" {
		return next
	}
	if q := strings.TrimSpace(r.URL.Query().Get(nextField)); q != "" {
		return SafePathNext(q, homeRoute)
	}
	return homeRoute
}

func clearLoginNextCookie(w http.ResponseWriter) {
	clearNamedCookie(w, loginNextCookie, loginRoute, http.SameSiteLaxMode)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, returnTo string) {
	setLoginNextCookie(w, returnTo)
	http.Redirect(w, r, loginRoute, http.StatusFound)
}

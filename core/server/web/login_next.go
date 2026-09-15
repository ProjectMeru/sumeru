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
	var next string
	if n := loginNextFromRequest(r); n != "" {
		next = n
	} else if q := strings.TrimSpace(r.URL.Query().Get(nextField)); q != "" {
		next = SafePathNext(q, homeRoute)
	} else {
		next = homeRoute
	}
	if uid := SessionUserID(r); uid > 0 {
		return postLoginDestination(r.Context(), uid, next)
	}
	return next
}

func clearLoginNextCookie(w http.ResponseWriter) {
	clearNamedCookie(w, loginNextCookie, loginRoute, http.SameSiteLaxMode)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, returnTo string) {
	setLoginNextCookie(w, returnTo)
	http.Redirect(w, r, loginRoute, http.StatusFound)
}

package web

import (
	"net/http"
)

func SessionTouchPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !sessionBrowserPresenceEnabled() {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if !ParsePostForm(w, r) || !validateSessionCSRF(w, r) || !requireLogin(w, r) {
		return
	}
	sid, _ := sessionCookieFromRequest(r)
	touchBrowserSeen(r.Context(), sid)
	w.WriteHeader(http.StatusNoContent)
}

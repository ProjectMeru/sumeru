package web

import (
	"net/http"
	"strings"
)

const apiKeyRevealRoute = "/web/apikey/reveal"

func registerAPIKeyRevealRoute() {
	registerSession(http.MethodGet, apiKeyRevealRoute, APIKeyRevealHandler)
}

// APIKeyRevealHandler shows a one-time API key from the flash cookie (never embedded in shell HTML).
func APIKeyRevealHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	raw := ConsumeAPIKeyFlash(r, w)
	if raw == "" {
		http.Error(w, "No API key pending display", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>API Key</title></head><body>`))
	_, _ = w.Write([]byte(`<h1>API key (shown once)</h1><p>Copy this key now. It will not be shown again.</p>`))
	_, _ = w.Write([]byte(`<pre style="user-select:all">`))
	_, _ = w.Write([]byte(htmlEscape(raw)))
	_, _ = w.Write([]byte(`</pre><p><a href="/web">Back to app</a></p></body></html>`))
}

func htmlEscape(s string) string {
	repl := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return repl.Replace(s)
}

package web

import (
	"net/http"
)

const apiKeyFlashCookie = "sumeru_api_key_flash"

// SetAPIKeyFlash stores a one-time raw API key in an HttpOnly cookie (never in redirect URLs).
func SetAPIKeyFlash(w http.ResponseWriter, raw string) {
	setNamedCookie(w, apiKeyFlashCookie, raw, "/", 120, http.SameSiteLaxMode)
}

// ConsumeAPIKeyFlash reads and clears the one-time API key flash cookie.
func ConsumeAPIKeyFlash(r *http.Request, w http.ResponseWriter) string {
	c, err := r.Cookie(apiKeyFlashCookie)
	clearNamedCookie(w, apiKeyFlashCookie, "/", http.SameSiteLaxMode)
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}

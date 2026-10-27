package web

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

const (
	apiKeyFlashCookie      = "sumeru_api_key_flash"
	recordErrorFlashCookie = "sumeru_record_error_flash"
)

type recordErrorFlashPayload struct {
	Kind        string   `json:"kind"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Details     string   `json:"details,omitempty"`
	FieldErrors []string `json:"field_errors,omitempty"`
}

// PageFlash is a one-time user-visible banner after redirect.
type PageFlash struct {
	Kind        string   `json:"kind"` // success, info, warning, error
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Details     string   `json:"details,omitempty"`
	FieldErrors []string `json:"field_errors,omitempty"`
}

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

func recordErrorFlashCookieAttrs() *http.Cookie {
	return buildNamedCookie(recordErrorFlashCookie, "", "/", 120, http.SameSiteLaxMode, true, sessionCookieSecure())
}

// SetRecordErrorFlash stores a one-time error banner in an HttpOnly cookie.
func SetRecordErrorFlash(w http.ResponseWriter, flash PageFlash) {
	payload, err := json.Marshal(recordErrorFlashPayload(flash))
	if err != nil {
		return
	}
	cookie := recordErrorFlashCookieAttrs()
	cookie.Value = base64.StdEncoding.EncodeToString(payload)
	setCookie(w, cookie)
}

// ConsumeRecordErrorFlash reads and clears the one-time record error flash cookie.
func ConsumeRecordErrorFlash(r *http.Request, w http.ResponseWriter) (PageFlash, bool) {
	clearNamedCookie(w, recordErrorFlashCookie, "/", http.SameSiteLaxMode)

	c, err := r.Cookie(recordErrorFlashCookie)
	if err != nil || c.Value == "" {
		return PageFlash{}, false
	}
	raw, err := base64.StdEncoding.DecodeString(c.Value)
	if err != nil {
		return PageFlash{}, false
	}
	var payload recordErrorFlashPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return PageFlash{}, false
	}
	if payload.Kind == "" && payload.Body == "" && payload.Title == "" {
		return PageFlash{}, false
	}
	return PageFlash(payload), true
}

// ConsumePageFlashes reads and clears one-time flash data (cookies).
func ConsumePageFlashes(r *http.Request, w http.ResponseWriter) []PageFlash {
	var out []PageFlash
	if c, err := r.Cookie(apiKeyFlashCookie); err == nil && c.Value != "" {
		out = append(out, PageFlash{
			Kind:  "success",
			Title: "API key created",
			Body:  "Open the one-time reveal page to copy your key (available for 2 minutes):\n/web/apikey/reveal",
		})
	}
	if flash, ok := ConsumeRecordErrorFlash(r, w); ok {
		out = append(out, flash)
	}
	return out
}

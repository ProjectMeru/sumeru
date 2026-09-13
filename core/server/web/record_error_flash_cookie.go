package web

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

const recordErrorFlashCookie = "sumeru_record_error_flash"

type recordErrorFlashPayload struct {
	Kind        string   `json:"kind"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Details     string   `json:"details,omitempty"`
	FieldErrors []string `json:"field_errors,omitempty"`
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
	http.SetCookie(w, cookie)
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

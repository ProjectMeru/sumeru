package web

import (
	"net/http"
)

// PageFlash is a one-time user-visible banner after redirect.
type PageFlash struct {
	Kind        string   `json:"kind"` // success, info, warning, error
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Details     string   `json:"details,omitempty"` // optional technical details for error flashes
	FieldErrors []string `json:"field_errors,omitempty"`
}

// ConsumePageFlashes reads and clears one-time flash data (cookies).
func ConsumePageFlashes(r *http.Request, w http.ResponseWriter) []PageFlash {
	var out []PageFlash
	if key := ConsumeAPIKeyFlash(r, w); key != "" {
		out = append(out, PageFlash{
			Kind:  "success",
			Title: "API key created",
			Body:  "Copy this key now — it will not be shown again:\n" + key,
		})
	}
	if flash, ok := ConsumeRecordErrorFlash(r, w); ok {
		out = append(out, flash)
	}
	return out
}

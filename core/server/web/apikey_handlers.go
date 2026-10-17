package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"sumeru/core/errcode"
	"sumeru/core/orm"
)

const apiKeyModel = "core.user.apikey"

const apiKeyRevealRoute = "/web/apikey/reveal"

func registerAPIKeyRevealRoute() {
	registerSession(http.MethodGet, apiKeyRevealRoute, APIKeyRevealHandler)
}

// ActionCreateAPIKey generates a one-time raw API key for a user.
func ActionCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	if !requireModelAccess(w, r, apiKeyModel, "create") {
		return
	}

	ctx := r.Context()
	targetUserID := apiKeyTargetUserID(r)
	if targetUserID <= 0 {
		http.Error(w, "Could not create API key", http.StatusUnauthorized)
		return
	}
	keyName := strings.TrimSpace(r.PostFormValue("name"))
	expiresAt, msg, err := apiKeyExpiresAtFromForm(r)
	if msg != "" {
		redirectWithWebMessage(w, r, r.PostFormValue("next"), msg)
		return
	}
	if err != nil {
		redirectWithWebMessage(w, r, r.PostFormValue("next"), "api_key_duration_invalid")
		return
	}

	rawKey, err := orm.CreateAPIKeyForUser(ctx, targetUserID, keyName, expiresAt)
	if err != nil {
		WebLogEvent(ctx, WebLogInput{
			Route:     createAPIKeyRoute,
			Message:   "Could not create API key",
			Code:      errcode.InternalError,
			Operation: "create_api_key",
			Status:    logStatusFailure,
			Err:       err,
			ContextFields: map[string]interface{}{
				"user_id": targetUserID,
			},
		})
		http.Error(w, "Could not create API key", http.StatusInternalServerError)
		return
	}

	SetAPIKeyFlash(w, rawKey)
	redirectWithWebMessage(w, r, r.PostFormValue("next"), "api_key_created")
}

func apiKeyTargetUserID(r *http.Request) int {
	sessionUID := SessionUserID(r)
	userID, _ := strconv.Atoi(strings.TrimSpace(r.PostFormValue("user_id")))
	if 0 >= userID || userID == sessionUID {
		return sessionUID
	}
	if orm.UserHasGroupXML(r.Context(), sessionUID, groupSystemXML) {
		return userID
	}
	return sessionUID
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

func apiKeyExpiresAtFromForm(r *http.Request) (time.Time, string, error) {
	daysStr := strings.TrimSpace(r.PostFormValue("duration_days"))
	if daysStr == "" {
		return time.Time{}, "api_key_duration_required", nil
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return time.Time{}, "api_key_duration_invalid", err
	}
	expires, err := orm.ExpiresAtFromDurationDays(days, time.Now().UTC())
	if err != nil {
		return time.Time{}, "api_key_duration_invalid", err
	}
	return expires, "", nil
}

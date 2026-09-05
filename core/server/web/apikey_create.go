package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/errcode"
	"sumeru/core/orm"
)

const apiKeyModel = "core.user.apikey"

// ActionCreateAPIKey generates a one-time raw API key for a user.
//
// Form fields: name (optional label), user_id (optional; defaults to signed-in user), next (redirect target).
// The raw key is never placed in the redirect URL — it is delivered once via SetAPIKeyFlash
// and shown on the next page by ConsumePageFlashes (see page_flash.go).
func ActionCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	if !requireModelAccess(w, r, apiKeyModel, "create") {
		return
	}

	ctx := r.Context()
	targetUserID := apiKeyTargetUserID(r)
	keyName := strings.TrimSpace(r.PostFormValue("name"))

	rawKey, err := orm.CreateAPIKeyForUser(ctx, targetUserID, keyName)
	if err != nil {
		WebLogEvent(ctx, WebLogInput{
			Route:     "/web/action/create_api_key",
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

// apiKeyTargetUserID resolves which user receives the new key; falls back to the session user.
// Targeting another user requires base.group_system.
func apiKeyTargetUserID(r *http.Request) int {
	sessionUID := SessionUserID(r)
	userID, _ := strconv.Atoi(strings.TrimSpace(r.PostFormValue("user_id")))
	if userID <= 0 || userID == sessionUID {
		return sessionUID
	}
	if orm.UserHasGroupXML(r.Context(), sessionUID, groupSystemXML) {
		return userID
	}
	return sessionUID
}

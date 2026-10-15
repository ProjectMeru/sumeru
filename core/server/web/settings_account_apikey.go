package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sumeru/core/errcode"
	"sumeru/core/orm"
)

type settingsAccountAPIKeyRow struct {
	ID         int
	Name       string
	Prefix     string
	ExpiresAt  string
	LastUsedAt string
	Active     bool
}

func loadSettingsAccountAPIKeys(ctx context.Context, userID int) []settingsAccountAPIKeyRow {
	if userID <= 0 || orm.DB == nil {
		return nil
	}
	rows, err := orm.SearchPage(ctx, "core.user.apikey", [][]interface{}{
		{"user_id", "=", userID},
	}, 50, 0, "id desc")
	if err != nil {
		return nil
	}
	out := make([]settingsAccountAPIKeyRow, 0, len(rows))
	for _, row := range rows {
		id64, ok := orm.CoerceInt64(row["id"])
		if !ok || id64 <= 0 {
			continue
		}
		name, _ := row["name"].(string)
		prefix, _ := row["key_prefix"].(string)
		active, _ := row["active"].(bool)
		out = append(out, settingsAccountAPIKeyRow{
			ID:         int(id64),
			Name:       strings.TrimSpace(name),
			Prefix:     strings.TrimSpace(prefix),
			ExpiresAt:  formatAPIKeyDisplayTime(row["expires_at"]),
			LastUsedAt: formatAPIKeyDisplayTime(row["last_used_at"]),
			Active:     active,
		})
	}
	return out
}

func formatAPIKeyDisplayTime(v interface{}) string {
	s, _ := v.(string)
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

func handleSettingsAccountAPIKey(w http.ResponseWriter, r *http.Request, ctx context.Context, actor int, action string) bool {
	switch action {
	case "create_api_key":
		handleSettingsAccountCreateAPIKey(w, r, ctx, actor)
		return true
	case "revoke_api_key":
		handleSettingsAccountRevokeAPIKey(w, r, ctx, actor)
		return true
	default:
		return false
	}
}

func handleSettingsAccountCreateAPIKey(w http.ResponseWriter, r *http.Request, ctx context.Context, actor int) {
	expiresAt, msg, err := apiKeyExpiresAtFromForm(r)
	if msg != "" {
		redirectWithWebMessage(w, r, settingsAccountRoute, msg)
		return
	}
	if err != nil {
		redirectWithWebMessage(w, r, settingsAccountRoute, "api_key_duration_invalid")
		return
	}
	name := strings.TrimSpace(r.PostForm.Get("api_key_name"))
	rawKey, err := orm.CreateAPIKeyForUser(ctx, actor, name, expiresAt)
	if err != nil {
		WebLogEvent(ctx, WebLogInput{
			Route:     settingsAccountRoute,
			Message:   "Could not create API key",
			Code:      errcode.InternalError,
			Operation: "create_api_key",
			Status:    logStatusFailure,
			Err:       err,
			ContextFields: map[string]interface{}{
				"user_id": actor,
			},
		})
		redirectWithWebMessage(w, r, settingsAccountRoute, "api_key_create_failed")
		return
	}
	SetAPIKeyFlash(w, rawKey)
	redirectWithWebMessage(w, r, settingsAccountRoute, "api_key_created")
}

func handleSettingsAccountRevokeAPIKey(w http.ResponseWriter, r *http.Request, ctx context.Context, actor int) {
	keyID, _ := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("key_id")))
	if keyID <= 0 {
		redirectWithWebMessage(w, r, settingsAccountRoute, "api_key_revoke_failed")
		return
	}
	row, err := orm.SearchOne(ctx, "core.user.apikey", map[string]interface{}{
		"id":      keyID,
		"user_id": actor,
	})
	if err != nil || row == nil {
		redirectWithWebMessage(w, r, settingsAccountRoute, "api_key_revoke_failed")
		return
	}
	if err := orm.UpdateRecordByID(ctx, "core.user.apikey", keyID, map[string]interface{}{"active": false}); err != nil {
		redirectWithWebMessage(w, r, settingsAccountRoute, "api_key_revoke_failed")
		return
	}
	redirectWithWebMessage(w, r, settingsAccountRoute, "api_key_revoked")
}

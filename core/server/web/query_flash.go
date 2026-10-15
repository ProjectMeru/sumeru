package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"sumeru/core/engine/render"
)

var importFlashPattern = regexp.MustCompile(`^imported_(\d+)_updated_(\d+)_skipped_(\d+)$`)

// FlashFromQueryMessage converts ?msg= query values into workspace flash banners.
func FlashFromQueryMessage(msg string) (render.FlashMessage, bool) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return render.FlashMessage{}, false
	}
	if m := importFlashPattern.FindStringSubmatch(msg); len(m) == 4 {
		created, _ := strconv.Atoi(m[1])
		updated, _ := strconv.Atoi(m[2])
		skipped, _ := strconv.Atoi(m[3])
		return render.FlashMessage{
			Kind:  "success",
			Title: "Import complete",
			Body:  fmt.Sprintf("Created %d, updated %d, skipped %d row(s).", created, updated, skipped),
		}, true
	}
	if strings.HasPrefix(msg, "imported_") {
		n := strings.TrimPrefix(msg, "imported_")
		return render.FlashMessage{
			Kind:  "success",
			Title: "Import complete",
			Body:  fmt.Sprintf("Imported %s row(s).", n),
		}, true
	}
	switch msg {
	case resetPasswordMsg:
		return render.FlashMessage{Kind: "info", Title: "Password reset", Body: "If the account exists, reset instructions were sent."}, true
	case oauthDeniedMsg:
		return render.FlashMessage{Kind: "error", Title: "Sign-in failed", Body: "Single sign-on could not complete. Try again or use your password if enabled."}, true
	case authLocalDisabledMsg:
		return render.FlashMessage{Kind: "error", Title: "Password sign-in disabled", Body: "Use one of the sign-in providers below."}, true
	case "totp_enroll_started":
		return render.FlashMessage{Kind: "info", Title: "Scan the QR code", Body: "Add the account in your authenticator app, then enter the 6-digit code to enable two-factor authentication."}, true
	case "totp_enabled":
		return render.FlashMessage{Kind: "success", Title: "Two-factor enabled", Body: "You will need an authenticator code when signing in on new devices."}, true
	case "totp_disabled":
		return render.FlashMessage{Kind: "success", Title: "Two-factor disabled", Body: "Authenticator codes are no longer required for your account."}, true
	case "totp_invalid":
		return render.FlashMessage{Kind: "error", Title: "Invalid code", Body: "Check the 6-digit code from your authenticator app and try again."}, true
	case "totp_enroll_failed":
		return render.FlashMessage{Kind: "error", Title: "Could not start enrollment", Body: "Try again or contact your administrator."}, true
	case "password_updated":
		return render.FlashMessage{Kind: "success", Title: "Password updated", Body: "Your password was changed."}, true
	case "password_mismatch":
		return render.FlashMessage{Kind: "error", Title: "Passwords do not match", Body: "Enter the same password in both fields."}, true
	case "password_required":
		return render.FlashMessage{Kind: "error", Title: "Password required", Body: "Enter a new password."}, true
	case "password_failed":
		return render.FlashMessage{Kind: "error", Title: "Could not update password", Body: "Check policy requirements and try again."}, true
	case "auth_settings_saved":
		return render.FlashMessage{Kind: "success", Title: "Authentication settings saved", Body: "System parameters were updated."}, true
	case "auth_settings_invalid":
		return render.FlashMessage{Kind: "error", Title: "Invalid settings", Body: "Check password minimum length (6–72) and try again."}, true
	case "auth_settings_failed":
		return render.FlashMessage{Kind: "error", Title: "Could not save settings", Body: "Try again or contact your administrator."}, true
	case "field_acl_saved":
		return render.FlashMessage{Kind: "success", Title: "Field access saved", Body: "Matrix rules were updated."}, true
	case "field_acl_failed":
		return render.FlashMessage{Kind: "error", Title: "Field access not saved", Body: "Check the model and try again."}, true
	case "model_acl_saved":
		return render.FlashMessage{Kind: "success", Title: "Model access saved", Body: "Matrix rules were updated."}, true
	case "model_acl_failed":
		return render.FlashMessage{Kind: "error", Title: "Model access not saved", Body: "Check the model and try again."}, true
	case "api_key_created":
		return render.FlashMessage{Kind: "success", Title: "API key created", Body: "Copy the key from the banner above if shown."}, true
	case moduleMsgSaved:
		return render.FlashMessage{Kind: "success", Title: "Saved", Body: "Changes were saved."}, true
	case saveOKCreatedMsg:
		return render.FlashMessage{Kind: "success", Title: "Saved", Body: "Record created."}, true
	case saveOKUpdatedMsg:
		return render.FlashMessage{Kind: "success", Title: "Saved", Body: "Changes saved."}, true
	case stageUpdatedMsg:
		return render.FlashMessage{Kind: "success", Title: "Updated", Body: "Stage updated.", ToastOnly: true}, true
	default:
		if strings.HasPrefix(msg, "error:") {
			body := sanitizeFlashQueryBody(strings.TrimPrefix(msg, "error:"))
			return render.FlashMessage{Kind: "error", Title: "Error", Body: body}, true
		}
		if strings.HasPrefix(msg, "save_error:") {
			body := sanitizeFlashQueryBody(strings.TrimPrefix(msg, "save_error:"))
			return render.FlashMessage{Kind: "error", Title: "Save failed", Body: body}, true
		}
		if strings.HasPrefix(msg, "installed_") || strings.HasPrefix(msg, "uninstalled_") || strings.HasPrefix(msg, "upgraded_") {
			return render.FlashMessage{Kind: "success", Title: "Apps", Body: strings.ReplaceAll(msg, "_", " ")}, true
		}
		return render.FlashMessage{Kind: "info", Title: "", Body: sanitizeFlashQueryBody(msg)}, true
	}
}

func sanitizeFlashQueryBody(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	const maxRunes = 500
	runes := []rune(s)
	if len(runes) > maxRunes {
		runes = runes[:maxRunes]
	}
	out := make([]rune, 0, len(runes))
	for _, r := range runes {
		if r == '\n' || r == '\r' {
			out = append(out, ' ')
			continue
		}
		if r < 0x20 {
			continue
		}
		out = append(out, r)
	}
	return strings.TrimSpace(string(out))
}

func appendQueryFlashesToViewRecord(r *http.Request, viewRecord *render.ViewRecordData) {
	if viewRecord == nil || r == nil {
		return
	}
	msg := strings.TrimSpace(r.URL.Query().Get(flashMessageParam))
	if msg == "" {
		return
	}
	if flash, ok := FlashFromQueryMessage(msg); ok {
		viewRecord.FlashMessages = append(viewRecord.FlashMessages, flash)
	}
}

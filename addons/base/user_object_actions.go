package base

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/mail"
	"sumeru/core/sdk"
)

func init() {
	sdk.RegisterObjectAction("core.user", "action_archive", archiveCoreUser)
	sdk.RegisterObjectAction("core.user", "action_reset_password", resetPasswordCoreUser)
}

func archiveCoreUser(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	_ = model
	if err := sdk.CheckModelAccess(ctx, sdk.SecurityUID(ctx), "core.user", "write"); err != nil {
		return "", err
	}
	if _, err := sdk.SearchOne(ctx, "core.user", map[string]interface{}{"id": id}); err != nil {
		return "", fmt.Errorf("record not found")
	}
	if _, err := sdk.Write(ctx, "core.user", []int{id}, map[string]interface{}{"active": false}); err != nil {
		return "", err
	}
	return "", nil
}

func resetPasswordCoreUser(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	_ = model
	_ = vals
	if !sdk.UserHasGroupXML(ctx, sdk.SecurityUID(ctx), "base.group_system") {
		return "", fmt.Errorf("access denied")
	}
	rec, err := sdk.SearchOne(ctx, "core.user", map[string]interface{}{"id": id})
	if err != nil {
		return "", fmt.Errorf("record not found")
	}
	login := sdk.AsString(rec["login"])
	to := strings.TrimSpace(sdk.AsString(rec["email"]))
	if to == "" && strings.Contains(login, "@") {
		to = login
	}
	if mail.Configured() && to != "" {
		token, err := sdk.IssuePasswordReset(ctx, id)
		if err != nil {
			return "", err
		}
		if err := mail.SendPasswordResetEmail(ctx, to, login, sdk.PasswordResetURL(token)); err != nil {
			return "", fmt.Errorf("send reset email: %w", err)
		}
		applog.InfoMsg(ctx, "base", "user_action", "Password reset email sent",
			map[string]interface{}{"user_id": id, "login": login, "to": to})
		return "", nil
	}
	applog.InfoMsg(ctx, "base", "user_action", "Password reset requested (configure smtp_host/smtp_from to send email)",
		map[string]interface{}{"user_id": id, "login": login})
	return "", nil
}

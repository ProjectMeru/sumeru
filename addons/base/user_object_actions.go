package base

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/mail"
	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("core.user", "action_archive", archiveCoreUser)
	orm.RegisterObjectAction("core.user", "action_reset_password", resetPasswordCoreUser)
}

func archiveCoreUser(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	_ = model
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), "core.user", "write"); err != nil {
		return "", err
	}
	if _, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": id}); err != nil {
		return "", fmt.Errorf("record not found")
	}
	if err := orm.UpdateRecordByID(ctx, "core.user", id, map[string]interface{}{"active": false}); err != nil {
		return "", err
	}
	return "", nil
}

func resetPasswordCoreUser(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	_ = model
	_ = vals
	if !orm.UserHasGroupXML(ctx, orm.SecurityUID(ctx), "base.group_system") {
		return "", fmt.Errorf("access denied")
	}
	rec, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": id})
	if err != nil {
		return "", fmt.Errorf("record not found")
	}
	login := orm.AsString(rec["login"])
	to := strings.TrimSpace(orm.AsString(rec["email"]))
	if to == "" && strings.Contains(login, "@") {
		to = login
	}
	loginURL := "/web/login"
	if mail.Configured() && to != "" {
		if err := mail.SendPasswordResetEmail(ctx, to, login, loginURL); err != nil {
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

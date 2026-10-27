package sdk

import (
	"context"

	"sumeru/core/orm"
)

// CheckModelAccess reports whether uid may perform op on model.
func CheckModelAccess(ctx context.Context, uid int, model, op string) error {
	return orm.CheckModelAccess(ctx, uid, model, op)
}

// SecurityUID returns the user id stored on ctx.
func SecurityUID(ctx context.Context) int { return orm.SecurityUID(ctx) }

// SecurityBypass reports whether ctx skips record rules.
func SecurityBypass(ctx context.Context) bool { return orm.SecurityBypass(ctx) }

// UserHasGroupXML reports whether uid has the group identified by xmlID.
func UserHasGroupXML(ctx context.Context, uid int, xmlID string) bool {
	return orm.UserHasGroupXML(ctx, uid, xmlID)
}

// ContextWithUID stores uid on a new context.
func ContextWithUID(ctx context.Context, uid int) context.Context {
	return orm.ContextWithUID(ctx, uid)
}

// ContextWithCompanyID stores the active company on a new context.
func ContextWithCompanyID(ctx context.Context, companyID int64) context.Context {
	return orm.ContextWithCompanyID(ctx, companyID)
}

// CompanyIDFromContext returns the active company id, or 0.
func CompanyIDFromContext(ctx context.Context) int64 { return orm.CompanyIDFromContext(ctx) }

// ActiveCompanyIDForUser returns the company id for uid.
func ActiveCompanyIDForUser(ctx context.Context, uid int) int64 {
	return orm.ActiveCompanyIDForUser(ctx, uid)
}

// IssuePasswordReset creates a one-time reset token for the user.
func IssuePasswordReset(ctx context.Context, id int) (string, error) {
	return orm.IssuePasswordReset(ctx, id)
}

// PasswordResetURL builds the public reset link for token.
func PasswordResetURL(token string) string { return orm.PasswordResetURL(token) }

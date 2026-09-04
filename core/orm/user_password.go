package orm

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"

	"golang.org/x/crypto/bcrypt"
)

type ctxKeyAllowPasswordHash struct{}

// ContextAllowPasswordHashWrite marks ctx so PrepareValues may accept core.user.password
// (bcrypt hash only). Used exclusively by SetUserPassword / SetUserPasswordHash.
func ContextAllowPasswordHashWrite(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyAllowPasswordHash{}, true)
}

func passwordHashWriteAllowed(ctx context.Context) bool {
	v, _ := ctx.Value(ctxKeyAllowPasswordHash{}).(bool)
	return v
}

// SetUserPassword hashes plain under policy and stores it. Requires system admin.
func SetUserPassword(ctx context.Context, actor, userID int, plain string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	if actor <= 0 {
		return fmt.Errorf("unauthenticated")
	}
	if !UserHasGroupXML(ctx, actor, "base.group_system") {
		applog.WarnMsg(ctx, "orm", "user_password", "deny password change: actor not system admin", nil,
			map[string]interface{}{"user_id": userID, "actor": actor})
		return fmt.Errorf("password change requires system administrator")
	}
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return fmt.Errorf("password required")
	}
	if err := ValidatePasswordPolicy(plain); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return SetUserPasswordHash(ctx, userID, string(hash))
}

// SetUserPasswordHash stores a pre-computed bcrypt hash (internal / trusted callers only).
func SetUserPasswordHash(ctx context.Context, userID int, hash string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return fmt.Errorf("password hash required")
	}
	ctx = ContextAllowPasswordHashWrite(ctx)
	return UpdateRecordByID(ctx, "core.user", userID, map[string]interface{}{"password": hash})
}

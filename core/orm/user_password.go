package orm

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/server/config"

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

// SetOwnUserPassword lets a user change their own password (no current-password check yet).
func SetOwnUserPassword(ctx context.Context, actor, userID int, plain string) error {
	if 0 >= userID || 0 >= actor {
		return fmt.Errorf("unauthenticated")
	}
	if actor != userID && !UserHasGroupXML(ctx, actor, "base.group_system") {
		return fmt.Errorf("password change requires system administrator")
	}
	if actor != userID {
		return SetUserPassword(ctx, actor, userID, plain)
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

// SetUserPassword hashes plain under policy and stores it. Requires system admin.
func SetUserPassword(ctx context.Context, actor, userID int, plain string) error {
	if 0 >= userID {
		return fmt.Errorf("invalid user id")
	}
	if 0 >= actor {
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
	if 0 >= userID {
		return fmt.Errorf("invalid user id")
	}
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return fmt.Errorf("password hash required")
	}
	ctx = ContextAllowPasswordHashWrite(ctx)
	if err := UpdateRecordByID(ctx, "core.user", userID, map[string]interface{}{"password": hash}); err != nil {
		return err
	}
	DestroySessionsForUser(ctx, userID)
	return nil
}

const passwordResetTTL = time.Hour

// IssuePasswordReset stores a hash of a one-time token and returns the raw token.
func IssuePasswordReset(ctx context.Context, userID int) (string, error) {
	if DB == nil || 0 >= userID {
		return "", fmt.Errorf("invalid user")
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])
	expires := time.Now().UTC().Add(passwordResetTTL)
	tbl := MustQuotedTableName("core.user")
	_, err := DB.ExecContext(ctx,
		`UPDATE `+tbl+` SET reset_token_hash = $1, reset_token_expires = $2 WHERE id = $3`,
		hash, expires, userID)
	if err != nil {
		return "", err
	}
	return token, nil
}

// PasswordResetURL is the public link that carries the raw reset token.
func PasswordResetURL(token string) string {
	path := "/web/action/reset_password?token=" + url.QueryEscape(token)
	base := strings.TrimRight(strings.TrimSpace(config.AppConfig.WebBaseURL), "/")
	if base == "" {
		return path
	}
	return base + path
}

// ConsumePasswordReset sets a new password when the token hash matches and has not expired.
func ConsumePasswordReset(ctx context.Context, token, password, confirm string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("reset link is invalid or expired")
	}
	if password != confirm {
		return fmt.Errorf("password confirmation does not match")
	}
	if err := ValidatePasswordPolicy(password); err != nil {
		return err
	}
	if DB == nil {
		return fmt.Errorf("reset link is invalid or expired")
	}
	sum := sha256.Sum256([]byte(token))
	want := hex.EncodeToString(sum[:])
	tbl := MustQuotedTableName("core.user")
	var id int
	var hash string
	var expires time.Time
	err := DB.QueryRowContext(ctx,
		`SELECT id, COALESCE(reset_token_hash, ''), reset_token_expires FROM `+tbl+` WHERE reset_token_hash = $1`,
		want).Scan(&id, &hash, &expires)
	if err != nil || !hmac.Equal([]byte(hash), []byte(want)) || expires.Before(time.Now().UTC()) {
		return fmt.Errorf("reset link is invalid or expired")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := SetUserPasswordHash(ctx, id, string(hashed)); err != nil {
		return err
	}
	_, err = DB.ExecContext(ctx, `UPDATE `+tbl+` SET reset_token_hash = '', reset_token_expires = NULL WHERE id = $1`, id)
	return err
}

func DestroySessionsForUser(ctx context.Context, userID int) {
	if DB == nil || 0 >= userID {
		return
	}
	tbl := MustQuotedTableName("sys.session")
	_, _ = DB.ExecContext(ctx, `DELETE FROM `+tbl+` WHERE user_id = $1`, userID)
}

package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"sumeru/core/server/auth"
)

// BeginOwnTOTPEnrollment stores a new secret with totp_enabled=false for the given user.
func BeginOwnTOTPEnrollment(ctx context.Context, actor, userID int) (secret string, err error) {
	if err := assertOwnUser(actor, userID); err != nil {
		return "", err
	}
	secret, err = auth.GenerateTOTPSecret()
	if err != nil {
		return "", err
	}
	return secret, writeUserTOTPSecret(ctx, userID, secret, false)
}

// ConfirmOwnTOTPEnrollment validates code against stored secret and enables TOTP.
func ConfirmOwnTOTPEnrollment(ctx context.Context, actor, userID int, code string) error {
	if err := assertOwnUser(actor, userID); err != nil {
		return err
	}
	secret, enabled, err := readUserTOTP(ctx, userID)
	if err != nil {
		return err
	}
	if enabled {
		return fmt.Errorf("two-factor authentication is already enabled")
	}
	if secret == "" {
		return fmt.Errorf("start enrollment before confirming")
	}
	if !auth.ValidateTOTP(secret, code) {
		return fmt.Errorf("invalid authentication code")
	}
	return writeUserTOTPSecret(ctx, userID, secret, true)
}

// DisableOwnTOTP requires a valid code and clears TOTP for the user.
func DisableOwnTOTP(ctx context.Context, actor, userID int, code string) error {
	if err := assertOwnUser(actor, userID); err != nil {
		return err
	}
	secret, enabled, err := readUserTOTP(ctx, userID)
	if err != nil {
		return err
	}
	if !enabled || secret == "" {
		return fmt.Errorf("two-factor authentication is not enabled")
	}
	if !auth.ValidateTOTP(secret, code) {
		return fmt.Errorf("invalid authentication code")
	}
	return writeUserTOTPSecret(ctx, userID, "", false)
}

// DisableUserTOTP clears TOTP for userID (system administrator).
func DisableUserTOTP(ctx context.Context, actor, userID int) error {
	if 0 >= userID {
		return fmt.Errorf("invalid user id")
	}
	if 0 >= actor {
		return fmt.Errorf("unauthenticated")
	}
	if !UserHasGroupXML(ctx, actor, "base.group_system") {
		return fmt.Errorf("two-factor reset requires system administrator")
	}
	return writeUserTOTPSecret(ctx, userID, "", false)
}

// UserTOTPEnabled reports whether TOTP login is required for userID.
func UserTOTPEnabled(ctx context.Context, userID int) (bool, error) {
	_, enabled, err := readUserTOTP(ctx, userID)
	return enabled, err
}

// UserTOTPCredentialsForLogin returns secret/enabled for MFA verification (kernel read).
func UserTOTPCredentialsForLogin(ctx context.Context, userID int) (secret string, enabled bool) {
	secret, enabled, err := readUserTOTP(ctx, userID)
	if err != nil {
		return "", false
	}
	return secret, enabled
}

// OwnTOTPEnrollmentPending returns the pending secret when enrollment started but not confirmed.
func OwnTOTPEnrollmentPending(ctx context.Context, actor, userID int) (secret string, ok bool, err error) {
	if err := assertOwnUser(actor, userID); err != nil {
		return "", false, err
	}
	secret, enabled, err := readUserTOTP(ctx, userID)
	if err != nil {
		return "", false, err
	}
	if enabled || secret == "" {
		return "", false, nil
	}
	return secret, true, nil
}

func assertOwnUser(actor, userID int) error {
	if 0 >= userID || 0 >= actor {
		return fmt.Errorf("unauthenticated")
	}
	if actor != userID {
		return fmt.Errorf("two-factor enrollment requires your own account")
	}
	return nil
}

func readUserTOTP(ctx context.Context, userID int) (secret string, enabled bool, err error) {
	if DB == nil {
		return "", false, fmt.Errorf("database unavailable")
	}
	bypass := AuditedBypass(ctx, "user.totp.read")
	tbl := MustQuotedTableName("core.user")
	var secretNull sql.NullString
	var enabledVal bool
	err = DB.QueryRowContext(bypass,
		`SELECT COALESCE(totp_secret, ''), COALESCE(totp_enabled, false) FROM `+tbl+` WHERE id = $1`,
		userID,
	).Scan(&secretNull, &enabledVal)
	if err == sql.ErrNoRows {
		return "", false, fmt.Errorf("user not found")
	}
	if err != nil {
		return "", false, err
	}
	return strings.TrimSpace(secretNull.String), enabledVal, nil
}

func writeUserTOTPSecret(ctx context.Context, userID int, secret string, enabled bool) error {
	values := map[string]interface{}{
		"totp_secret":  strings.TrimSpace(secret),
		"totp_enabled": enabled,
	}
	return WithElevated(ctx, "user.totp.write", func(bypass context.Context) error {
		return UpdateRecordByID(bypass, "core.user", userID, values)
	})
}

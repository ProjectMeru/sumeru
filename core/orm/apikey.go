package orm

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// HashAPIKey returns a hex SHA-256 digest of the raw key.
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// GenerateAPIKey creates a random key sk_<hex>. Returns raw key (show once), prefix, and hash.
func GenerateAPIKey() (raw, prefix, hash string, err error) {
	buf := make([]byte, 24)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	raw = "sk_" + hex.EncodeToString(buf)
	prefix = raw
	if len(prefix) > 11 {
		prefix = prefix[:11]
	}
	return raw, prefix, HashAPIKey(raw), nil
}

// UIDFromAPIKey resolves an active API key to a user id. Returns 0 if invalid.
func UIDFromAPIKey(ctx context.Context, raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" || DB == nil {
		return 0
	}
	hash := HashAPIKey(raw)
	tbl := MustQuotedTableName("core.user.apikey")
	var uid int
	var active bool
	var expiresAt sql.NullString
	err := DB.QueryRowContext(ctx,
		`SELECT user_id, active, expires_at FROM `+tbl+` WHERE key_hash = $1 LIMIT 1`, hash,
	).Scan(&uid, &active, &expiresAt)
	if err == sql.ErrNoRows || err != nil || !active || uid <= 0 {
		return 0
	}
	if expiresAt.Valid && strings.TrimSpace(expiresAt.String) != "" {
		if t, parseErr := time.Parse(time.RFC3339, expiresAt.String); parseErr == nil && time.Now().UTC().After(t) {
			return 0
		}
	}
	_, _ = DB.ExecContext(ctx, `UPDATE `+tbl+` SET last_used_at = $1 WHERE key_hash = $2`, time.Now().UTC().Format(time.RFC3339), hash)
	// Ensure user is active.
	var userActive bool
	err = DB.QueryRowContext(ctx,
		`SELECT active FROM `+MustQuotedTableName("core.user")+` WHERE id = $1`, uid,
	).Scan(&userActive)
	if err != nil || !userActive {
		return 0
	}
	return uid
}

// CreateAPIKeyForUser inserts a new key row. Returns the raw key (caller must show once).
func CreateAPIKeyForUser(ctx context.Context, userID int, name string) (raw string, err error) {
	if userID <= 0 {
		return "", fmt.Errorf("user id required")
	}
	inst, ok := Registry["core.user.apikey"]
	if !ok {
		return "", fmt.Errorf("unknown model %q", "core.user.apikey")
	}
	raw, prefix, hash, err := GenerateAPIKey()
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		if seq, seqErr := NextSequence(ctx, "core.user.apikey"); seqErr == nil {
			name = seq
		} else {
			name = "API Key"
		}
	}
	_, err = Create(ctx, inst, map[string]interface{}{
		"user_id":     userID,
		"name":        name,
		"key_prefix":  prefix,
		"key_hash":    hash,
		"active":      true,
		"create_date": time.Now().UTC().Format(time.RFC3339),
	})
	return raw, err
}

// AppendUserLog writes a login audit row (best-effort).
func AppendUserLog(ctx context.Context, userID int, ip, result string) {
	if DB == nil {
		return
	}
	inst, ok := Registry["core.user.log"]
	if !ok {
		return
	}
	result = strings.TrimSpace(result)
	if result == "" {
		result = "success"
	}
	vals := map[string]interface{}{
		"create_date": time.Now().UTC().Format(time.RFC3339),
		"ip":          strings.TrimSpace(ip),
		"result":      result,
	}
	if userID > 0 {
		vals["user_id"] = userID
	}
	bypass := ContextWithBypass(ctx, true)
	_, _ = Create(bypass, inst, vals)
}

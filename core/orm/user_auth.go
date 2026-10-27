package orm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when login or password does not match.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserInactive is returned when the user record is inactive.
	ErrUserInactive = errors.New("user inactive")
)

const defaultSessionHours = 24

var loginDummyHash = mustLoginDummyBcryptHash()

func mustLoginDummyBcryptHash() string {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(h)
}

// AuthenticateLocalUser checks login and active state, then bcrypt-compares the stored password hash.
// Search redacts password, so the hash written at setup is read directly.
func AuthenticateLocalUser(ctx context.Context, login, password string) (uid int, username, name string, err error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return 0, "", "", ErrInvalidCredentials
	}
	if DB == nil {
		comparePasswordHash("", password)
		return 0, "", "", ErrInvalidCredentials
	}
	ctx = AuditedBypass(ctx, "auth.local")
	rows, err := SearchPage(ctx, "core.user", [][]interface{}{{"login", "ilike", login}}, 5, 0, "")
	if err != nil {
		return 0, "", "", err
	}
	var match map[string]interface{}
	for _, row := range rows {
		l, _ := row["login"].(string)
		if strings.EqualFold(strings.TrimSpace(l), login) {
			match = row
			break
		}
	}
	if match == nil {
		comparePasswordHash("", password)
		return 0, "", "", ErrInvalidCredentials
	}
	id64, ok := CoerceInt64(match["id"])
	if !ok || id64 <= 0 {
		return 0, "", "", ErrInvalidCredentials
	}
	id := int(id64)
	active, _ := match["active"].(bool)
	if !active {
		comparePasswordHash("", password)
		return 0, "", "", ErrInvalidCredentials
	}
	var hash string
	userTbl := MustQuotedTableName("core.user")
	if err := DB.QueryRowContext(ctx, `SELECT COALESCE(password, '') FROM `+userTbl+` WHERE id = $1`, id).Scan(&hash); err != nil {
		return 0, "", "", err
	}
	if !comparePasswordHash(hash, password) {
		return 0, "", "", ErrInvalidCredentials
	}
	username, _ = match["login"].(string)
	name, _ = match["name"].(string)
	if strings.TrimSpace(name) == "" {
		name = username
	}
	return id, username, name, nil
}

// SessionCreateOpts controls optional sys.session fields at create time.
type SessionCreateOpts struct {
	StampBrowserSeen bool
}

// CreateUserSession inserts sys.session via ORM Create and returns the sid.
func CreateUserSession(ctx context.Context, userID int, opts SessionCreateOpts) (sessionID string, err error) {
	if userID <= 0 {
		return "", fmt.Errorf("user id required")
	}
	inst, ok := Registry["sys.session"]
	if !ok {
		return "", fmt.Errorf("unknown model %q", "sys.session")
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	sessionID = hex.EncodeToString(buf)
	expiresAt := time.Now().UTC().Add(defaultSessionHours * time.Hour).Format(time.RFC3339)
	vals := map[string]interface{}{
		"sid":        sessionID,
		"user_id":    userID,
		"expires_at": expiresAt,
	}
	if opts.StampBrowserSeen {
		vals["browser_seen_at"] = time.Now().UTC().Format(time.RFC3339)
	}
	ctx = AuditedBypass(ctx, "session.create")
	if _, err := Create(ctx, inst, vals); err != nil {
		return "", err
	}
	return sessionID, nil
}

// IssueAPIKeyForLogin mints an API key for auth.login (24h TTL, elevated create).
func IssueAPIKeyForLogin(ctx context.Context, userID int) (raw string, err error) {
	ctx = AuditedBypass(ctx, "auth.login.apikey")
	expires := time.Now().UTC().Add(AuthLoginAPIKeyLifetime)
	return CreateAPIKeyForUser(ctx, userID, "auth.login", expires)
}

// ResolveCompanyTenant resolves X-Sumeru-Tenant to a company id (numeric or name).
func ResolveCompanyTenant(ctx context.Context, tenantHeader string) (companyID int64, err error) {
	tenantHeader = strings.TrimSpace(tenantHeader)
	if tenantHeader == "" {
		return 0, fmt.Errorf("tenant required")
	}
	if id, parseErr := strconv.ParseInt(tenantHeader, 10, 64); parseErr == nil && id > 0 {
		return id, nil
	}
	ctx = AuditedBypass(ctx, "tenant.resolve")
	row, err := SearchOne(ctx, "core.company", map[string]interface{}{"name": tenantHeader})
	if err != nil {
		return 0, fmt.Errorf("unknown tenant %q", tenantHeader)
	}
	cid, ok := CoerceInt64(row["id"])
	if !ok || cid <= 0 {
		return 0, fmt.Errorf("unknown tenant %q", tenantHeader)
	}
	return cid, nil
}

func comparePasswordHash(storedHash, plain string) bool {
	storedHash = strings.TrimSpace(storedHash)
	if storedHash == "" {
		_ = bcrypt.CompareHashAndPassword([]byte(loginDummyHash), []byte(plain))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plain)) == nil
}

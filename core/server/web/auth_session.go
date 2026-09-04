package web

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const sessionCookieName = "sumeru_session"
const sessionDuration = 7 * 24 * time.Hour
const sessionSlidingTTL = 24 * time.Hour

var testSessionUserIDOverride int

func randomSessionID() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func buildSessionCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !config.AppConfig.DevMode,
	}
}

// CreateSession stores a new session and sets an HttpOnly cookie.
func CreateSession(w http.ResponseWriter, userID int) error {
	if orm.DB == nil {
		return fmt.Errorf("no database")
	}
	sessionID, err := randomSessionID()
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(sessionDuration)
	sessionTable := orm.MustQuotedTableName("sys.session")
	if _, err := orm.DB.Exec(`INSERT INTO `+sessionTable+` (sid, user_id, expires_at) VALUES ($1, $2, $3)`, sessionID, userID, expiresAt); err != nil {
		return err
	}
	http.SetCookie(w, buildSessionCookie(sessionID, int(sessionDuration.Seconds())))
	return nil
}

// ClearSessionCookie removes the session cookie (client-side).
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, buildSessionCookie("", -1))
}

// SessionUserID returns core.user id from cookie session, or 0.
func SessionUserID(r *http.Request) int {
	if testSessionUserIDOverride > 0 {
		return testSessionUserIDOverride
	}
	if orm.DB == nil {
		return 0
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return 0
	}
	sessionTable := orm.MustQuotedTableName("sys.session")
	var userID int
	err = orm.DB.QueryRow(
		`SELECT user_id FROM `+sessionTable+` WHERE sid = $1 AND expires_at > NOW()`,
		cookie.Value,
	).Scan(&userID)
	if err != nil {
		return 0
	}
	// Sliding idle expiry (DB-backed; works across instances).
	_, _ = orm.DB.Exec(
		`UPDATE `+sessionTable+` SET expires_at = $1 WHERE sid = $2 AND expires_at > NOW()`,
		time.Now().UTC().Add(sessionSlidingTTL),
		cookie.Value,
	)
	return userID
}

// DestroySession removes the session row and clears the cookie.
func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		sessionTable := orm.MustQuotedTableName("sys.session")
		_, _ = orm.DB.Exec(`DELETE FROM `+sessionTable+` WHERE sid = $1`, cookie.Value)
	}
	ClearSessionCookie(w)
}

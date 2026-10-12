package web

import (
	"context"
	"database/sql"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const defaultSessionBrowserGraceSeconds = 90

func sessionBrowserPresenceEnabled() bool {
	return config.AppConfig.CloseSessionOnLastTab
}

func sessionBrowserGrace() time.Duration {
	sec := config.AppConfig.SessionBrowserGraceSeconds
	if sec <= 0 {
		sec = defaultSessionBrowserGraceSeconds
	}
	return time.Duration(sec) * time.Second
}

func touchBrowserSeen(ctx context.Context, sid string) {
	if orm.DB == nil || sid == "" || !sessionBrowserPresenceEnabled() {
		return
	}
	sessionTbl := orm.MustQuotedTableName("sys.session")
	_, _ = orm.DB.ExecContext(ctx,
		`UPDATE `+sessionTbl+` SET browser_seen_at = NOW() WHERE sid = $1 AND expires_at > NOW()`,
		sid,
	)
}

func browserSeenExpired(seen sql.NullTime, now time.Time) bool {
	if !sessionBrowserPresenceEnabled() {
		return false
	}
	if !seen.Valid {
		return false
	}
	return now.Sub(seen.Time) > sessionBrowserGrace()
}

package orm

import (
	"context"
)

// DestroySessionsForUser deletes all DB-backed sessions for a user.
func DestroySessionsForUser(ctx context.Context, userID int) {
	if DB == nil || userID <= 0 {
		return
	}
	sessionTable := MustQuotedTableName("sys.session")
	_, _ = DB.ExecContext(ctx, `DELETE FROM `+sessionTable+` WHERE user_id = $1`, userID)
}

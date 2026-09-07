//go:build integration

package orm_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"sumeru/core/orm"
)

const testPasswordHash = "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"

func integrationCtx() context.Context {
	ctx := orm.ContextWithBypass(context.Background(), true)
	return orm.ContextWithUID(ctx, 1)
}

func initIntegrationDB(t *testing.T) context.Context {
	t.Helper()
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDBWithPool(dsn, orm.DBPoolSettings{MaxOpenConns: 5, MaxIdleConns: 2})
	if !orm.IsInitialized() {
		t.Skip("database not initialized")
	}
	return integrationCtx()
}

func createTestUser(t *testing.T, ctx context.Context) int {
	t.Helper()
	userModel, ok := orm.Registry["core.user"]
	if !ok {
		t.Fatal("core.user not registered")
	}
	login := fmt.Sprintf("sess_revoke_%d", time.Now().UnixNano())
	userID, err := orm.Create(ctx, userModel, map[string]interface{}{
		"login": login,
		"name":  "Session Revoke Test",
	})
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return userID
}

func insertSessions(t *testing.T, ctx context.Context, userID, count int) {
	t.Helper()
	sessionTable := orm.MustQuotedTableName("sys.session")
	for i := 0; i < count; i++ {
		sid := fmt.Sprintf("test-sid-%d-%d-%d", userID, i, time.Now().UnixNano())
		_, err := orm.DB.ExecContext(ctx,
			`INSERT INTO `+sessionTable+` (sid, user_id, expires_at) VALUES ($1, $2, NOW() + interval '1 day')`,
			sid, userID,
		)
		if err != nil {
			t.Fatalf("insert session: %v", err)
		}
	}
}

func sessionCount(t *testing.T, ctx context.Context, userID int) int {
	t.Helper()
	sessionTable := orm.MustQuotedTableName("sys.session")
	var count int
	err := orm.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+sessionTable+` WHERE user_id = $1`, userID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	return count
}

func TestPasswordChangeRevokesSessions(t *testing.T) {
	ctx := initIntegrationDB(t)
	userID := createTestUser(t, ctx)
	insertSessions(t, ctx, userID, 2)
	if sessionCount(t, ctx, userID) != 2 {
		t.Fatal("expected 2 sessions before password change")
	}
	if err := orm.SetUserPasswordHash(ctx, userID, testPasswordHash); err != nil {
		t.Fatalf("SetUserPasswordHash: %v", err)
	}
	if got := sessionCount(t, ctx, userID); got != 0 {
		t.Fatalf("expected 0 sessions after password change, got %d", got)
	}
}

func TestDeactivateUserRevokesSessions(t *testing.T) {
	ctx := initIntegrationDB(t)
	userID := createTestUser(t, ctx)
	insertSessions(t, ctx, userID, 1)
	if sessionCount(t, ctx, userID) != 1 {
		t.Fatal("expected 1 session before deactivation")
	}
	if err := orm.UpdateRecordByID(ctx, "core.user", userID, map[string]interface{}{"active": false}); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
	if got := sessionCount(t, ctx, userID); got != 0 {
		t.Fatalf("expected 0 sessions after deactivation, got %d", got)
	}
}

//go:build integration

package module_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"sumeru/core/module"
	"sumeru/core/orm"
)

func integrationDB(t *testing.T) context.Context {
	t.Helper()
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	preflight, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := preflight.Ping(); err != nil {
		_ = preflight.Close()
		t.Skipf("database unavailable: %v", err)
	}
	_ = preflight.Close()

	orm.InitDBWithPool(dsn, orm.DBPoolSettings{MaxOpenConns: 5, MaxIdleConns: 2})
	if !orm.IsInitialized() {
		t.Skip("database schema not bootstrapped (run sumeru -i base)")
	}
	return orm.ContextWithBypass(context.Background(), true)
}

func TestReconcileSystemManagerImpliedInstalled_linksCRMManager(t *testing.T) {
	ctx := integrationDB(t)
	if _, err := orm.SearchOne(ctx, "sys.module", map[string]interface{}{"name": "crm", "state": "installed"}); err != nil {
		t.Skip("crm module not installed")
	}
	sysGID, _, err := orm.ResolveXmlId(ctx, "base.group_system")
	if err != nil || sysGID <= 0 {
		t.Fatalf("resolve base.group_system: %v", err)
	}
	mgrGID, _, err := orm.ResolveXmlId(ctx, "crm.group_crm_manager")
	if err != nil || mgrGID <= 0 {
		t.Skip("crm.group_crm_manager not in database")
	}

	module.ReconcileSystemManagerImpliedInstalled(ctx)

	implTbl := orm.MustQuotedTableName("core.group.implied")
	var count int
	err = orm.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+implTbl+` WHERE group_id = $1 AND implied_group_id = $2`,
		sysGID, mgrGID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query implied: %v", err)
	}
	if count < 1 {
		t.Fatal("expected base.group_system to imply crm.group_crm_manager after reconcile")
	}
}

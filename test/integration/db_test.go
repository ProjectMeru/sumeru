//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	_ "sumeru/core/ormmodels"
)

func TestIntegrationDBPing(t *testing.T) {
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDBWithPool(dsn, orm.DBPoolSettings{MaxOpenConns: 5, MaxIdleConns: 2})
	if !orm.IsInitialized() {
		t.Skip("database not initialized (run setup or apply schema first)")
	}
	_ = modelreg.ActivateAll(nil)
	ctx, cancel := context.WithTimeout(orm.ContextWithBypass(context.Background(), true), 10*time.Second)
	defer cancel()
	if _, err := orm.SearchLimit(ctx, "sys.module", nil, 5); err != nil {
		t.Fatalf("search sys.module: %v", err)
	}
}

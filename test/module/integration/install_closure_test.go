//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	_ "sumeru/core/ormmodels"
)

func TestIntegrationModuleRegistrySearchable(t *testing.T) {
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDB(dsn)
	_ = modelreg.ActivateAll(nil)
	ctx := orm.ContextWithBypass(context.Background(), true)
	if _, err := orm.SearchLimit(ctx, "sys.module", nil, 5); err != nil {
		t.Fatalf("search sys.module: %v", err)
	}
}

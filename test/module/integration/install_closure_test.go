//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"sumeru/core/orm"
)

func TestIntegrationModuleRegistrySearchable(t *testing.T) {
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	if err := orm.InitDB(dsn); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	ctx := context.Background()
	if _, err := orm.SearchLimit(ctx, "sys.module", nil, 5); err != nil {
		t.Fatalf("search sys.module: %v", err)
	}
}

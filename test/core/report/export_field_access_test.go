//go:build integration

package report_test

import (
	"context"
	"os"
	"testing"

	_ "sumeru/addons/base"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/report"
)

func TestExportCSVRejectsFieldReadDenyWhenUIDSet(t *testing.T) {
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	orm.InitDB(dsn)
	_ = modelreg.ActivateAll(nil)
	if !orm.IsInitialized() {
		t.Skip("database not bootstrapped")
	}
	ctx := orm.ContextWithUID(context.Background(), 2)
	_, err := report.ExportCSVUnlimitedForTest(ctx, report.ExportCSVInput{
		Model:  "core.user",
		Fields: []string{"login", "password"},
	})
	if err == nil {
		t.Fatal("expected field access error when exporting password for non-superuser")
	}
}

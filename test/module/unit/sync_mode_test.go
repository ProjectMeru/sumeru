package module_test

import (
	"context"
	"testing"

	"sumeru/core/module"
)

func TestDataFileOptsSkipExistingOnUpdate(t *testing.T) {
	opts := module.NewDataFileOptsForTest(true)
	ctx := module.ContextWithSyncMode(context.Background(), module.ModuleReloadUpdateForTest)
	if opts.SkipExistingOnUpdateForTest(ctx, "base", "") {
		t.Fatal("empty xml id should not skip")
	}
	ctxInstall := module.ContextWithSyncMode(context.Background(), module.ModuleReloadInstallForTest)
	if opts.SkipExistingOnUpdateForTest(ctxInstall, "base", "company_main") {
		t.Fatal("install mode should not skip")
	}
	if opts.SkipExistingOnUpdateForTest(ctx, "base", "company_main") {
		t.Fatal("without DB/xml id resolution should not skip when id unknown")
	}
	opts2 := module.NewDataFileOptsForTest(false)
	if opts2.SkipExistingOnUpdateForTest(ctx, "base", "company_main") {
		t.Fatal("noUpdate false should not skip")
	}
}

func TestSyncModeFromContextDefault(t *testing.T) {
	if module.SyncModeFromContext(context.Background()) != module.ModuleReloadInstallForTest {
		t.Fatal("expected install default")
	}
}

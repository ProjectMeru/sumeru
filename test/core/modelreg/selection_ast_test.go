package modelreg_test

import (
	"path/filepath"
	"testing"

	"sumeru/core/modelreg"
	"sumeru/test/harness"
)

func TestSelectionConstsFromSourcePackage(t *testing.T) {
	root := harness.RepoRoot(t)
	pkgDir := filepath.Join(root, "core", "modelreg", "testdata", "selection")
	opts := modelreg.SelectionOptionsFromPackageForTest(pkgDir, "Priority")
	if len(opts) != 3 {
		t.Fatalf("expected 3 priority options, got %v", opts)
	}
}

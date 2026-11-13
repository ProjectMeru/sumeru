package security_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/test/harness"
)

// TestElevationReasonsInKernel documents WithElevated call sites (SEC-021); CI still enforces check_security_bypass.sh.
func TestElevationReasonsInKernel(t *testing.T) {
	root := harness.RepoRoot(t)
	checks := []struct {
		file   string
		needle string
	}{
		{"core/scheduler/scheduler.go", `WithElevated(ctx, "cron.run"`},
		{"core/module/install.go", `WithElevated(ctx, "module.install"`},
		{"core/orm/outbox_drain.go", `WithElevated(ctx, "outbox.drain"`},
	}
	var missing []string
	for _, c := range checks {
		data, err := os.ReadFile(filepath.Join(root, c.file))
		if err != nil {
			missing = append(missing, c.file)
			continue
		}
		if !strings.Contains(string(data), c.needle) {
			missing = append(missing, c.file+": "+c.needle)
		}
	}
	if 0 < len(missing) {
		t.Fatalf("elevation inventory missing: %v", missing)
	}
}

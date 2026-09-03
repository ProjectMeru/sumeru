package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"sumeru/test/harness"
)

func TestSumeruModuleUsage(t *testing.T) {
	root := harness.RepoRoot(t)
	cmd := exec.Command("go", "run", "./cmd/sumeru-module")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for missing command")
	}
	if len(out) == 0 {
		t.Fatal("expected usage output")
	}
}

func TestSumeruModuleListMissingConfig(t *testing.T) {
	root := harness.RepoRoot(t)
	cmd := exec.Command("go", "run", "./cmd/sumeru-module", "-c", filepath.Join(t.TempDir(), "missing.conf"), "list")
	cmd.Dir = root
	if err := cmd.Run(); err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestSumeruDbCheckMissingConfig(t *testing.T) {
	root := harness.RepoRoot(t)
	cmd := exec.Command("go", "run", "./cmd/sumeru-db-check", "-c", filepath.Join(t.TempDir(), "missing.conf"))
	cmd.Dir = root
	if err := cmd.Run(); err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestSumeruShellHelp(t *testing.T) {
	root := harness.RepoRoot(t)
	cmd := exec.Command("go", "run", "./cmd/sumeru-shell", "-h")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, out)
	}
}

func TestSumeruImportGenHelp(t *testing.T) {
	root := harness.RepoRoot(t)
	cmd := exec.Command("go", "run", "./cmd/sumeru-import-gen", "-h")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("help failed: %v\n%s", err, out)
	}
}

func TestSumeruMainMissingConfig(t *testing.T) {
	if os.Getenv("SUMERU_CMD_SMOKE") == "0" {
		t.Skip("cmd smoke disabled")
	}
	root := harness.RepoRoot(t)
	cmd := exec.Command("go", "run", "./cmd/sumeru", "-c", filepath.Join(t.TempDir(), "missing.conf"), "--help")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("help failed: %v\n%s", err, out)
	}
}

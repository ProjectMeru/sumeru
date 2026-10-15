package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/test/harness"
)

func TestSumeruBpScaffoldsI18n(t *testing.T) {
	root := harness.RepoRoot(t)
	out := t.TempDir()
	cmd := exec.Command("go", "run", "./cmd/sumeru-bp", "-name", "bp_i18n_demo", "-out", out)
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("scaffold: %v\n%s", err, output)
	}
	catalog, err := os.ReadFile(filepath.Join(out, "bp_i18n_demo", "i18n", "en_US.po"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(catalog)
	if !strings.Contains(text, "Language: en_US") || !strings.Contains(text, "Bp I18n Demo") {
		t.Fatalf("catalog missing source strings:\n%s", text)
	}
	again := exec.Command("go", "run", "./cmd/sumeru-bp", "-name", "bp_i18n_demo", "-out", out)
	again.Dir = root
	if err := again.Run(); err == nil {
		t.Fatal("expected non-zero exit when the module directory already exists")
	}
}

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

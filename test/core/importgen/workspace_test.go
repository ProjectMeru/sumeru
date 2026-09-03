package importgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/core/importgen"
)

func TestRunWorkspaceGen_scopedWrites(t *testing.T) {
	sumeruRoot := testRepoRoot(t)
	addonsRoot := filepath.Join(filepath.Dir(sumeruRoot), "sumeru_addons")
	if _, err := os.Stat(addonsRoot); err != nil {
		t.Skipf("sumeru_addons not found at %s", addonsRoot)
	}

	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "go.mod"), "module test_workspace\n\ngo 1.22\n")
	writeFile(t, filepath.Join(workspace, "sumeru.conf"), strings.Join([]string{
		"[options]",
		"db_host = localhost",
		"db_port = 5432",
		"db_user = postgres",
		"db_password = postgres",
		"db_name = test",
		"http_port = 8080",
		"sumeru_home = " + sumeruRoot,
		"addons_path = " + filepath.Join(sumeruRoot, "addons") + "," + addonsRoot + "," + filepath.Join(workspace, "addons"),
	}, "\n")+"\n")

	addon := filepath.Join(workspace, "addons", "demo")
	writeFile(t, filepath.Join(addon, "manifest.json"), `{"name":"demo","version":"1.0.0","author":"test","description":"test","depends":["base"]}`)
	writeFile(t, filepath.Join(addon, "models", "models.go"), `package models


type Demo struct {
	sdk.Model
	Name string
}
`)

	sentinel := filepath.Join(sumeruRoot, "addons", "base", "models", "zmodels.go")
	before, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("read sentinel: %v", err)
	}

	out := filepath.Join(workspace, "addonimports", "zimports.go")
	dest, err := importgen.RunWorkspaceGen(workspace, sumeruRoot, addonsRoot, "sumeru.conf", out, "addonimports")
	if err != nil {
		t.Fatal(err)
	}
	if dest != filepath.Clean(out) {
		t.Fatalf("dest: got %q want %q", dest, filepath.Clean(out))
	}

	after, err := os.ReadFile(sentinel)
	if err != nil {
		t.Fatalf("read sentinel after: %v", err)
	}
	if string(after) != string(before) {
		t.Fatal("RunWorkspaceGen modified sumeru addon zmodels.go")
	}

	for _, path := range []string{
		filepath.Join(addon, "init.go"),
		filepath.Join(addon, "models", "zmodels.go"),
		out,
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

package importgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"sumeru/core/importgen"
	"sumeru/core/module"
)


func testRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := module.FindRepoRoot(wd)
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}
	return root
}

func TestValidateOutputPackage(t *testing.T) {
	if err := importgen.ValidateOutputPackage("main"); err != nil {
		t.Fatal(err)
	}
	if err := importgen.ValidateOutputPackage("addonimports"); err != nil {
		t.Fatal(err)
	}
	if err := importgen.ValidateOutputPackage(""); err == nil {
		t.Fatal("expected error for empty package")
	}
	if err := importgen.ValidateOutputPackage("Bad"); err == nil {
		t.Fatal("expected error for uppercase")
	}
	if err := importgen.ValidateOutputPackage("bad-pkg"); err == nil {
		t.Fatal("expected error for hyphen")
	}
}

func TestResolveOutputPath(t *testing.T) {
	repo := "/opt/sumeru"
	if got, err := importgen.ResolveOutputPath(repo, "cmd/sumeru/zimports.go"); err != nil {
		t.Fatal(err)
	} else {
		want := filepath.Clean(filepath.Join(repo, "cmd/sumeru/zimports.go"))
		if got != want {
			t.Fatalf("relative: got %q want %q", got, want)
		}
	}
	abs := filepath.Join(t.TempDir(), "nested", "z.go")
	if got, err := importgen.ResolveOutputPath(repo, abs); err != nil {
		t.Fatal(err)
	} else if got != filepath.Clean(abs) {
		t.Fatalf("absolute: got %q want %q", got, filepath.Clean(abs))
	}
}

func TestBuildImportGoFile(t *testing.T) {
	s := importgen.BuildImportGoFile("addonimports", []string{"sumeru/addons/sales", "sumeru/addons/crm"})
	if !strings.Contains(s, "package addonimports") {
		t.Fatalf("missing package: %s", s)
	}
	if !strings.Contains(s, `_ "sumeru/addons/sales"`) {
		t.Fatal("missing import")
	}
}

func TestRunGen_addonimportsExternalOut(t *testing.T) {
	repoRoot := testRepoRoot(t)
	out := filepath.Join(t.TempDir(), "addonimports", "zimports.go")
	dest, err := importgen.RunGen(repoRoot, "sumeru.conf.example", out, "addonimports")
	if err != nil {
		t.Fatal(err)
	}
	if dest != filepath.Clean(out) {
		t.Fatalf("dest: got %q want %q", dest, filepath.Clean(out))
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	if !strings.Contains(body, "package addonimports") {
		t.Fatalf("body: %s", body)
	}
	if !strings.Contains(body, `_ "sumeru/addons/`) && !strings.Contains(body, `_ "sumeru/core/sdk/`) {
		t.Fatalf("expected sumeru blank imports, got: %s", body)
	}
}

func TestRunGen_defaultPackageMain(t *testing.T) {
	repoRoot := testRepoRoot(t)
	out := filepath.Join(t.TempDir(), "zimports_main.go")
	_, err := importgen.RunGen(repoRoot, "sumeru.conf.example", out, "main")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "package main") {
		t.Fatalf("expected package main: %s", b)
	}
}

func TestRunGen_invalidPackage(t *testing.T) {
	repoRoot := testRepoRoot(t)
	out := filepath.Join(t.TempDir(), "x.go")
	_, err := importgen.RunGen(repoRoot, "sumeru.conf.example", out, "NotValid")
	if err == nil {
		t.Fatal("expected error")
	}
}

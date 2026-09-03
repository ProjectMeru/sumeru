package importgen_test

import (
	"path/filepath"
	"strings"
	"testing"

	"sumeru/core/importgen"
	"sumeru/core/module"
)

func writeGoMod(t *testing.T, dir, modulePath string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "go.mod"), "module "+modulePath+"\n\ngo 1.22\n")
}

func TestRenderAddonInitIncludesDepends(t *testing.T) {
	addonsRoot := filepath.Join(t.TempDir(), "sumeru_addons")
	writeGoMod(t, addonsRoot, "sumeru_addons")
	crmPath := filepath.Join(addonsRoot, "crm")
	salePath := filepath.Join(addonsRoot, "sale")
	saleCrmPath := filepath.Join(addonsRoot, "sale_crm")
	writeFile(t, filepath.Join(saleCrmPath, "models", "models.go"), "package models\n")

	discovered := map[string]*module.Addon{
		"crm": {
			Manifest: module.Manifest{Name: "crm", Depends: []string{"base"}},
			Path:     crmPath,
		},
		"sale": {
			Manifest: module.Manifest{Name: "sale", Depends: []string{"base"}},
			Path:     salePath,
		},
		"sale_crm": {
			Manifest: module.Manifest{Name: "sale_crm", Depends: []string{"crm", "sale"}},
			Path:     saleCrmPath,
		},
	}
	imports, err := importgen.CollectInitImportsForTest(discovered, discovered["sale_crm"])
	if err != nil {
		t.Fatal(err)
	}
	body := importgen.RenderAddonInitForTest("sale_crm", imports)
	if !strings.Contains(body, "package sale_crm") {
		t.Fatalf("missing package: %s", body)
	}
	crmIdx := strings.Index(body, "sumeru_addons/crm\"")
	saleIdx := strings.Index(body, "sumeru_addons/sale\"")
	modelsIdx := strings.Index(body, "sale_crm/models\"")
	if crmIdx < 0 || saleIdx < 0 || modelsIdx < 0 {
		t.Fatalf("missing expected imports: %s", body)
	}
	if crmIdx >= modelsIdx || saleIdx >= modelsIdx {
		t.Fatalf("depends should precede local subpackages: %s", body)
	}
	if strings.Contains(body, "sumeru/addons/base\"") {
		t.Fatalf("direct depends only: should not import transitive base: %s", body)
	}
}

func TestExpectedInitDependsImportPathsWorkspace(t *testing.T) {
	root := testRepoRoot(t)
	workspace := t.TempDir()
	addonsRoot := filepath.Join(workspace, "sumeru_addons")
	customRoot := filepath.Join(workspace, "sumeru_custom_addons")
	writeGoMod(t, addonsRoot, "sumeru_addons")
	writeGoMod(t, customRoot, "sumeru_custom_addons")
	hrPath := filepath.Join(addonsRoot, "hr")
	cookbookPath := filepath.Join(customRoot, "addons", "engagement_cookbook")

	discovered := map[string]*module.Addon{
		"base": {
			Manifest: module.Manifest{Name: "base", Depends: []string{}},
			Path:     filepath.Join(root, "addons", "base"),
		},
		"contacts": {
			Manifest: module.Manifest{Name: "contacts", Depends: []string{"base"}},
			Path:     filepath.Join(root, "addons", "contacts"),
		},
		"mail": {
			Manifest: module.Manifest{Name: "mail", Depends: []string{"base"}},
			Path:     filepath.Join(root, "addons", "mail"),
		},
		"hr": {
			Manifest: module.Manifest{Name: "hr", Depends: []string{"base", "contacts"}},
			Path:     hrPath,
		},
		"engagement_cookbook": {
			Manifest: module.Manifest{
				Name:    "engagement_cookbook",
				Depends: []string{"base", "contacts", "hr", "mail"},
			},
			Path: cookbookPath,
		},
	}
	paths, err := module.ExpectedInitDependsImportPaths(discovered, discovered["engagement_cookbook"])
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 4 {
		t.Fatalf("expected 4 direct depend imports, got %v", paths)
	}
	set := map[string]struct{}{}
	for _, p := range paths {
		set[p] = struct{}{}
	}
	for _, want := range []string{
		"sumeru/addons/base",
		"sumeru/addons/contacts",
		"sumeru_addons/hr",
		"sumeru/addons/mail",
	} {
		if _, ok := set[want]; !ok {
			t.Fatalf("missing import %q in %v", want, paths)
		}
	}
}

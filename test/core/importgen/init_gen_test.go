package importgen_test

import (
	"strings"
	"testing"

	"sumeru/core/importgen"
	"sumeru/core/module"
)

func TestRenderAddonInitIncludesDepends(t *testing.T) {
	discovered := map[string]*module.Addon{
		"crm": {
			Manifest: module.Manifest{Name: "crm", Depends: []string{"base"}},
			Path:     testRepoRoot(t) + "/../sumeru_addons/crm",
		},
		"sale": {
			Manifest: module.Manifest{Name: "sale", Depends: []string{"base"}},
			Path:     testRepoRoot(t) + "/../sumeru_addons/sale",
		},
		"sale_crm": {
			Manifest: module.Manifest{Name: "sale_crm", Depends: []string{"crm", "sale"}},
			Path:     testRepoRoot(t) + "/../sumeru_addons/sale_crm",
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
	discovered := map[string]*module.Addon{
		"base": {
			Manifest: module.Manifest{Name: "base", Depends: []string{}},
			Path:     root + "/addons/base",
		},
		"contacts": {
			Manifest: module.Manifest{Name: "contacts", Depends: []string{"base"}},
			Path:     root + "/addons/contacts",
		},
		"mail": {
			Manifest: module.Manifest{Name: "mail", Depends: []string{"base"}},
			Path:     root + "/addons/mail",
		},
		"hr": {
			Manifest: module.Manifest{Name: "hr", Depends: []string{"base", "contacts"}},
			Path:     root + "/../sumeru_addons/hr",
		},
		"engagement_cookbook": {
			Manifest: module.Manifest{
				Name:    "engagement_cookbook",
				Depends: []string{"base", "contacts", "hr", "mail"},
			},
			Path: root + "/../sumeru_custom_addons/addons/engagement_cookbook",
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

package importgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/core/importgen"
)

func TestRenderZRefsForTest_table(t *testing.T) {
	t.Parallel()
	body := importgen.RenderZRefsForTest([]importgen.ExportedRefForTest{
		{Name: "SaleOrder", Kind: "phantom", TechnicalModel: "sale.order"},
		{Name: "CrmLead", Kind: "alias", TechnicalModel: "crm.lead", ImportPath: "sumeru/addons/crm/models", ImportAlias: "crmmodels", SourceGoName: "Lead"},
	})
	for _, want := range []string{"sale.order", "crm.lead", "package models", "sdk.Model"} {
		if !strings.Contains(body, want) {
			t.Fatalf("RenderZRefsForTest missing %q:\n%s", want, body)
		}
	}
}

func TestParseDirModelsForTest_tempDir(t *testing.T) {
	dir := t.TempDir()
	modelFile := filepath.Join(dir, "partner.go")
	content := `package models

type ResPartner struct {
	modelmeta.ModelMeta ` + "`sumeru:\"model=res.partner\"`" + `
	Name modelmeta.String
}
`
	if err := os.WriteFile(modelFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	models, err := importgen.ParseDirModelsForTest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) == 0 {
		t.Fatal("ParseDirModelsForTest found no models")
	}
}

func TestScanPackageForModelsForTest_missingDir(t *testing.T) {
	_, err := importgen.ScanPackageForModelsForTest(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected error for missing dir")
	}
}

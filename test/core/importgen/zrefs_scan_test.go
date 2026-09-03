package importgen_test

import (
	"sumeru/core/importgen"
	"os"
	"path/filepath"
	"testing"
)



func TestModelSpecFromStructAST(t *testing.T) {
	dir := t.TempDir()
	src := `package models


type CorePartner struct {
	modelmeta.ModelMeta ` + "`sumeru:\"model=core.partner\"`" + `
	Name string
}

type PartnerExt struct {
	modelmeta.ModelMeta ` + "`sumeru:\"inherit=core.partner\"`" + `
	Phone string
}

type Skipped struct {
	modelmeta.ModelMeta ` + "`sumeru:\"model=-\"`" + `
}

type DefaultName struct {
	modelmeta.ModelMeta
	Title string
}
`
	if err := os.WriteFile(filepath.Join(dir, "partner.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	models, err := importgen.ParseDirModelsForTest(dir)
	if err != nil {
		t.Fatalf("parseDirModels: %v", err)
	}
	byName := map[string]importgen.ScannedModelForTest{}
	for _, m := range models {
		byName[m.GoName] = m
	}

	if m := byName["CorePartner"]; m.ModelName != "core.partner" || m.Extend {
		t.Fatalf("CorePartner: %+v", m)
	}
	if m := byName["PartnerExt"]; m.ModelName != "core.partner" || !m.Extend {
		t.Fatalf("PartnerExt: %+v", m)
	}
	if m := byName["Skipped"]; m.ModelName != "-" {
		t.Fatalf("Skipped: %+v", m)
	}
	if m := byName["DefaultName"]; m.ModelName != "default.name" || m.Extend {
		t.Fatalf("DefaultName: %+v", m)
	}

	types, err := importgen.ScanPackageForModelsForTest(dir)
	if err != nil {
		t.Fatalf("scanPackageForModels: %v", err)
	}
	for _, name := range types {
		if name == "Skipped" {
			t.Fatal("model=- should be excluded from zmodels scan")
		}
	}
	if len(types) != 3 {
		t.Fatalf("expected 3 registered types, got %v", types)
	}
}

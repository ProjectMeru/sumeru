package modelreg_test

import (
	"path/filepath"
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/test/harness"
)

func TestMustRegisterManyFieldTypes(t *testing.T) {
	modelreg.ResetActivationForTest()
	orm.Registry = map[string]orm.Model{}

	type tagModel struct {
		modelmeta.ModelMeta `sumeru:"model=test.tags"`
		Active              modelmeta.Boolean
		Note                modelmeta.Text
		Amount              modelmeta.Numeric
		When                modelmeta.Date
		At                  modelmeta.DateTime
		Ref                 modelmeta.Many2One[struct{}] `sumeru:"comodel=res.partner"`
		Lines               modelmeta.One2Many[struct{}] `sumeru:"comodel=test.line"`
		Tags                modelmeta.Many2Many[struct{}] `sumeru:"comodel=res.tag"`
	}

	modelreg.MustRegister("test", &tagModel{})
	if err := modelreg.ActivateAll(nil); err != nil {
		t.Fatal(err)
	}
	m := orm.Registry["test.tags"]
	if m == nil {
		t.Fatal("missing model")
	}
	fields := map[string]orm.FieldType{}
	for _, f := range m.Fields() {
		fields[f.Name] = f.Type
	}
	for name, want := range map[string]orm.FieldType{
		"active": orm.Boolean,
		"note":   orm.Text,
		"amount": orm.Numeric,
		"when":   orm.Date,
		"at":     orm.DateTime,
	} {
		if fields[name] != want {
			t.Fatalf("%s type=%s want=%s", name, fields[name], want)
		}
	}
}

func TestSelectionOptionsFromPackage(t *testing.T) {
	root := harness.RepoRoot(t)
	opts := modelreg.SelectionOptionsFromPackageForTest(filepath.Join(root, "test/core/modelreg"), "testPriority")
	_ = opts
}

func TestParseSelectionEdgeCases(t *testing.T) {
	if got := modelreg.ParseSelectionForTest("only"); len(got) != 1 || got[0][1] != "Only" {
		t.Fatalf("single: %v", got)
	}
	if got := modelreg.ParseSelectionForTest("a:A,b:"); len(got) != 2 {
		t.Fatalf("empty label: %v", got)
	}
}

func TestParseDefaultEdgeCases(t *testing.T) {
	if modelreg.ParseDefaultForTest("", orm.Char) != "" {
		t.Fatal("empty default")
	}
	if modelreg.ParseDefaultForTest("true", orm.Boolean) != true {
		t.Fatal("bool true")
	}
	if modelreg.ParseDefaultForTest("false", orm.Boolean) != false {
		t.Fatal("bool false")
	}
}

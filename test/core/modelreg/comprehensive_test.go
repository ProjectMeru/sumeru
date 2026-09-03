package modelreg_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"sumeru/core/modelmeta"
	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/test/harness"
)

func TestMapMarkerType_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		marker string
		ft     orm.FieldType
		widget string
		err    bool
	}{
		{"String", orm.Char, "", false},
		{"Text", orm.Text, "", false},
		{"Boolean", orm.Boolean, "", false},
		{"Integer", orm.Integer, "", false},
		{"Float", orm.Float, "float", false},
		{"Date", orm.Date, "date", false},
		{"DateTime", orm.DateTime, "datetime", false},
		{"Selection", orm.Selection, "", false},
		{"NotARealType", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.marker, func(t *testing.T) {
			ft, widget, err := modelreg.MapMarkerTypeForTest(tc.marker)
			if tc.err {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil || ft != tc.ft || widget != tc.widget {
				t.Fatalf("MapMarkerTypeForTest(%q) = (%s, %q, %v)", tc.marker, ft, widget, err)
			}
		})
	}
}

func TestParseDefault_table(t *testing.T) {
	t.Parallel()
	tests := []struct {
		raw  string
		ft   orm.FieldType
		want interface{}
	}{
		{"hello", orm.Char, "hello"},
		{"3", orm.Integer, int64(3)},
		{"1.5", orm.Float, 1.5},
		{"low", orm.Selection, "low"},
	}
	for _, tc := range tests {
		if got := modelreg.ParseDefaultForTest(tc.raw, tc.ft); got != tc.want {
			t.Fatalf("ParseDefaultForTest(%q, %s) = %v want %v", tc.raw, tc.ft, got, tc.want)
		}
	}
}

func TestParseSelection_table(t *testing.T) {
	t.Parallel()
	got := modelreg.ParseSelectionForTest("draft:Draft,done:Done,cancelled:")
	if len(got) != 3 || got[2][1] != "Cancelled" {
		t.Fatalf("ParseSelectionForTest: %v", got)
	}
}

func TestRegisterCtxResolveComodel(t *testing.T) {
	t.Parallel()
	type Partner struct{}
	ctx := modelreg.NewRegisterCtxForTest().
		SetTypeMapping("Many2One", "res.partner", "Partner", reflect.TypeOf(Partner{}))
	comodel, err := ctx.ResolveComodel(reflect.TypeOf(Partner{}), modelmeta.FieldTags{Comodel: "res.partner"})
	if err != nil || comodel != "res.partner" {
		t.Fatalf("ResolveComodel: %q err=%v", comodel, err)
	}
}

func TestSelectionOptionsFromPackage_testdata(t *testing.T) {
	root := harness.RepoRoot(t)
	opts := modelreg.SelectionOptionsFromPackageForTest(
		filepath.Join(root, "core", "modelreg", "testdata", "selection"),
		"Priority",
	)
	if len(opts) != 3 {
		t.Fatalf("SelectionOptionsFromPackageForTest: %v", opts)
	}
}

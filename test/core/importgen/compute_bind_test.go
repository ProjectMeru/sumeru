package importgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sumeru/core/importgen"
)

func TestRenderPackageModelsBindsCompute(t *testing.T) {
	dir := t.TempDir()
	src := "package models\n\n" +
		"type EngagementDeliverable struct {\n" +
		"\tsdk.Model `sumeru:\"model=engagement.deliverable\"`\n" +
		"\tLineTotal sdk.Numeric `sumeru:\"compute=line_total,string=Line Total\"`\n" +
		"}\n"
	if err := os.WriteFile(filepath.Join(dir, "deliverable.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	body, err := importgen.RenderPackageModelsForTest(dir, "engagement_cookbook", "models")
	if err != nil {
		t.Fatal(err)
	}
	want := `sdk.BindCompute("engagement.deliverable", "line_total", line_total)`
	if !strings.Contains(body, want) {
		t.Fatalf("missing bind:\n%s", body)
	}
}

func TestRenderPackageModelsRejectsFormula(t *testing.T) {
	dir := t.TempDir()
	src := "package models\n\n" +
		"type EngagementDeliverable struct {\n" +
		"\tsdk.Model `sumeru:\"model=engagement.deliverable\"`\n" +
		"\tLineTotal sdk.Numeric `sumeru:\"compute=quantity * unit_price\"`\n" +
		"}\n"
	if err := os.WriteFile(filepath.Join(dir, "deliverable.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := importgen.RenderPackageModelsForTest(dir, "engagement_cookbook", "models")
	if err == nil || !strings.Contains(err.Error(), "not a function name") {
		t.Fatalf("got %v", err)
	}
}

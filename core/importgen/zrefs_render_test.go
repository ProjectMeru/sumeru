package importgen

import (
	"strings"
	"testing"
)

func TestRenderZRefs_phantomOmitsUnusedImport(t *testing.T) {
	body := renderZRefs([]exportedRef{{
		Name:           "CorePartner",
		Kind:           "phantom",
		TechnicalModel: "core.partner",
		ImportPath:     "sumeru/addons/contacts/models",
		ImportAlias:    "contactsmodels",
	}})
	if strings.Contains(body, "contactsmodels") {
		t.Fatalf("phantom ref must not import source package:\n%s", body)
	}
	if !strings.Contains(body, "sdk.Model `sumeru:\"model=core.partner\"`") {
		t.Fatalf("expected phantom stub:\n%s", body)
	}
}

func TestRenderZRefs_aliasIncludesImport(t *testing.T) {
	body := renderZRefs([]exportedRef{{
		Name:           "CorePartner",
		Kind:           "alias",
		TechnicalModel: "core.partner",
		ImportPath:     "sumeru/addons/base/models",
		ImportAlias:    "basemodels",
		SourceGoName:   "CorePartner",
	}})
	if !strings.Contains(body, `basemodels "sumeru/addons/base/models"`) {
		t.Fatalf("alias ref must import source package:\n%s", body)
	}
	if !strings.Contains(body, "type CorePartner = basemodels.CorePartner") {
		t.Fatalf("expected type alias:\n%s", body)
	}
}

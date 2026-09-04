package swcmeta_test

import (
	"testing"

	"sumeru/core/engine/swcmeta"
)

// ponytail: manual mirror in core/swc/src/types/workspace.ts — fail CI when Go payload fields drift.
func TestWorkspacePayloadJSONFieldsContract(t *testing.T) {
	want := []string{
		"actionId", "menuId", "viewType", "model", "recordId", "formEdit", "csrfToken",
		"arch", "record", "records", "viewTabs", "breadcrumbs", "listSearch", "listSearchUrl",
		"listTotal", "listSort", "listOffset", "listFilter", "listDomain", "listGroupBy",
		"listSections", "favorites", "formBaseQuery", "defaults", "iframeUrl",
	}
	typ := swcmeta.WorkspacePayloadTypeForTest()
	got := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := tag
		for j, c := range tag {
			if c == ',' {
				name = tag[:j]
				break
			}
		}
		got = append(got, name)
	}
	if len(got) != len(want) {
		t.Fatalf("field count drift: got %d want %d\n got: %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("json tag[%d] = %q want %q", i, got[i], want[i])
		}
	}
}

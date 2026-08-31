package render

import "testing"

func TestViewModeFilterSet(t *testing.T) {
	if viewModeFilterSet(nil) != nil {
		t.Fatal("nil modes should not filter")
	}
	if viewModeFilterSet([]string{}) != nil {
		t.Fatal("empty modes should not filter")
	}
	set := viewModeFilterSet([]string{" Map ", "list", "form"})
	if len(set) != 3 {
		t.Fatalf("got %d entries", len(set))
	}
	for _, mode := range []string{"map", "list", "form"} {
		if _, ok := set[mode]; !ok {
			t.Fatalf("missing %q", mode)
		}
	}
	if _, ok := set["kanban"]; ok {
		t.Fatal("kanban should not be allowed")
	}
}

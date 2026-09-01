package web

import (
	"testing"
)

func TestPartitionListSections(t *testing.T) {
	rows := []map[string]interface{}{
		{"id": 1, "state": "draft", "name": "A"},
		{"id": 2, "state": "done", "name": "B"},
		{"id": 3, "state": "draft", "name": "C"},
	}
	sections := partitionListSections(rows, "state")
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if sections[0].Count+sections[1].Count != 3 {
		t.Fatalf("row count mismatch")
	}
}

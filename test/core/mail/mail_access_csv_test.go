package mail_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestMailAccessCSV_noPermissiveEmptyGroupID(t *testing.T) {
	t.Helper()
	root := findRepoRoot(t)
	path := filepath.Join(root, "addons", "mail", "security", "sys.access.csv")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if len(rows) < 2 {
		t.Fatal("expected header and data rows")
	}
	header := rows[0]
	groupCol := csvColIndex(header, "group_id")
	if groupCol < 0 {
		t.Fatal("group_id column missing")
	}
	permCols := []string{"perm_write", "perm_create", "perm_unlink"}
	permIdx := make([]int, len(permCols))
	for i, name := range permCols {
		permIdx[i] = csvColIndex(header, name)
		if permIdx[i] < 0 {
			t.Fatalf("column %q missing", name)
		}
	}

	for _, row := range rows[1:] {
		if len(row) <= groupCol {
			continue
		}
		groupID := strings.TrimSpace(row[groupCol])
		if groupID != "" {
			continue
		}
		for i, col := range permIdx {
			if col >= len(row) {
				continue
			}
			on, err := strconv.Atoi(strings.TrimSpace(row[col]))
			if err != nil {
				t.Fatalf("parse %s: %v", permCols[i], err)
			}
			if on == 1 {
				t.Fatalf("row %q has empty group_id with %s=1", row[0], permCols[i])
			}
		}
	}
}

func csvColIndex(header []string, name string) int {
	for i, h := range header {
		if strings.TrimSpace(h) == name {
			return i
		}
	}
	return -1
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "addons", "mail")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find sumeru module root")
		}
		dir = parent
	}
}

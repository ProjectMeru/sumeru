package security_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNoUnsafeContentDispositionInWebHandlers(t *testing.T) {
	root := webRoot(t)
	pattern := regexp.MustCompile(`Content-Disposition",\s*"attachment;\s*filename=`)
	var hits []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if pattern.Match(data) {
			hits = append(hits, path)
		}
		return nil
	})
	if len(hits) > 0 {
		t.Fatalf("unsafe Content-Disposition concatenation in: %v", hits)
	}
}

func TestNoRawSetCookieOutsideHelpers(t *testing.T) {
	root := webRoot(t)
	allowlist := map[string]bool{
		filepath.Join(root, "cookie_helpers.go"): true,
	}
	pattern := regexp.MustCompile(`http\.SetCookie\(`)
	var hits []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if allowlist[path] {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if pattern.Match(data) {
			hits = append(hits, path)
		}
		return nil
	})
	if len(hits) > 0 {
		t.Fatalf("http.SetCookie outside cookie_helpers.go: %v", hits)
	}
}

func TestNoEmptyGroupIDACLRowsInAddons(t *testing.T) {
	addons := filepath.Join(moduleRoot(t), "addons")
	pattern := regexp.MustCompile(`,,\d,\d`)
	var hits []string
	_ = filepath.Walk(addons, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "sys.access.csv") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "id,") {
				continue
			}
			if pattern.MatchString(line) {
				hits = append(hits, path+": "+line)
			}
		}
		return nil
	})
	if len(hits) > 0 {
		t.Fatalf("empty group_id ACL rows: %v", hits)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func webRoot(t *testing.T) string {
	return filepath.Join(moduleRoot(t), "core", "server", "web")
}

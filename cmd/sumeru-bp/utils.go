package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// goStructName converts sale_order to SaleOrder.
func goStructName(technicalName string) string {
	return joinTitleWords(technicalName, "")
}

// displayName converts sale_order to Sale Order.
func displayName(technicalName string) string {
	return joinTitleWords(technicalName, " ")
}

func joinTitleWords(technicalName, separator string) string {
	parts := strings.Split(technicalName, "_")
	words := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		words = append(words, strings.ToUpper(lower[:1])+lower[1:])
	}
	return strings.Join(words, separator)
}

// modelsImportPath is the Go import of models/. A relative parent stays
// repo-relative. An absolute parent is made relative to the repo when it
// sits inside it; otherwise the import matches addons/<module>/models.
func modelsImportPath(goModule, repoRoot, parentDir, moduleName string) string {
	if !filepath.IsAbs(parentDir) {
		return goModule + "/" + filepath.ToSlash(filepath.Join(parentDir, moduleName, "models"))
	}
	rel, err := filepath.Rel(repoRoot, filepath.Join(parentDir, moduleName, "models"))
	outside := err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
	if outside {
		return goModule + "/addons/" + moduleName + "/models"
	}
	return goModule + "/" + filepath.ToSlash(rel)
}

func writeFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

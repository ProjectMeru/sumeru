package render

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/module"
	"sumeru/core/orm"
)

// RootMenuIDForModule returns the root sys.menu id for an installed module (parent_id IS NULL).
func RootMenuIDForModule(ctx context.Context, moduleName string) int {
	if orm.DB == nil || strings.TrimSpace(moduleName) == "" {
		return 0
	}
	table := orm.MustQuotedTableName("sys.menu")
	query := `SELECT id FROM ` + table + ` WHERE module = $1 AND parent_id IS NULL ORDER BY sequence ASC, id ASC LIMIT 1`
	var id int
	if err := orm.DB.QueryRowContext(ctx, query, strings.TrimSpace(moduleName)).Scan(&id); err != nil {
		return 0
	}
	return id
}

// RootMenuWebIconForModule returns the sanitized web_icon sprite key on a module root menu.
func RootMenuWebIconForModule(ctx context.Context, moduleName string) string {
	if orm.DB == nil || strings.TrimSpace(moduleName) == "" {
		return ""
	}
	table := orm.MustQuotedTableName("sys.menu")
	query := `SELECT COALESCE(NULLIF(TRIM(web_icon), ''), '') FROM ` + table +
		` WHERE module = $1 AND parent_id IS NULL ORDER BY sequence ASC, id ASC LIMIT 1`
	var icon string
	if err := orm.DB.QueryRowContext(ctx, query, strings.TrimSpace(moduleName)).Scan(&icon); err != nil {
		return ""
	}
	icon = strings.TrimSpace(icon)
	if menuIconKey.MatchString(icon) {
		return icon
	}
	return ""
}

// ModuleIconServePath returns the on-disk path for a module icon, or empty when unavailable.
func ModuleIconServePath(moduleName, iconRel string) string {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return ""
	}
	a := module.LoadedAddons[moduleName]
	if a == nil || a.Path == "" {
		return ""
	}
	root := filepath.Clean(a.Path)
	candidates := []string{}
	if iconRel = strings.TrimSpace(iconRel); iconRel != "" {
		candidates = append(candidates, iconRel)
	}
	candidates = append(candidates, "static/icon.png")
	for _, rel := range candidates {
		rel = strings.TrimSpace(rel)
		if rel == "" || strings.Contains(rel, `\`) || strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
			continue
		}
		if filepath.IsAbs(rel) || (len(rel) >= 2 && rel[1] == ':') {
			continue
		}
		rel = filepath.Clean(rel)
		if rel == "." || strings.HasPrefix(rel, "..") {
			continue
		}
		full := filepath.Join(root, rel)
		relToRoot, err := filepath.Rel(root, full)
		if err != nil || relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
			continue
		}
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			return full
		}
	}
	return ""
}

// ModuleIconURL returns /static/module-icon/<module> when an icon file exists on disk.
func ModuleIconURL(ctx context.Context, moduleName string) string {
	if orm.DB == nil || strings.TrimSpace(moduleName) == "" {
		return ""
	}
	row, err := orm.SearchOne(ctx, "sys.module", map[string]interface{}{"name": strings.TrimSpace(moduleName)})
	if err != nil {
		return ""
	}
	iconRel := strings.TrimSpace(orm.AsString(row["icon"]))
	if ModuleIconServePath(moduleName, iconRel) == "" {
		return ""
	}
	return "/static/module-icon/" + strings.TrimSpace(moduleName)
}

// IconLetterFromName returns the first letter of displayName for app launcher tiles.
func IconLetterFromName(displayName string) string {
	if runes := []rune(strings.TrimSpace(displayName)); len(runes) > 0 {
		return strings.ToUpper(string(runes[0]))
	}
	return "?"
}

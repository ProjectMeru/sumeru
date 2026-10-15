package orm

import (
	"context"
	"sort"
	"strings"
)

// PlatformModules materialize on a fresh DB and cannot be uninstalled.
var PlatformModules = map[string]bool{
	"base": true,
	"im":   true,
}

// IsPlatformModule reports whether name is part of the always-on platform spine.
func IsPlatformModule(name string) bool {
	return PlatformModules[name]
}

// DeclaringModule returns the sys.module technical name that owns this model's tables.
// Values are set at registration time via RegisterModelWithModule.
func DeclaringModule(modelName string) string {
	if m, ok := modelDeclaringModule[modelName]; ok {
		return strings.TrimSpace(m)
	}
	return ""
}

// InstalledModuleNames returns technical names of installed, active modules.
func InstalledModuleNames(ctx context.Context) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	if DB == nil {
		return out, nil
	}
	tbl := MustQuotedTableName("sys.module")
	rows, err := DB.QueryContext(ctx, `SELECT name FROM `+tbl+` WHERE state = 'installed' AND active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		n = strings.TrimSpace(n)
		if n != "" {
			out[n] = struct{}{}
		}
	}
	return out, rows.Err()
}

// ModelsOwnedByModule lists registered model names whose DeclaringModule matches moduleName.
func ModelsOwnedByModule(moduleName string) []string {
	var names []string
	for n := range Registry {
		if DeclaringModule(n) == moduleName {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

// ShouldMaterializeModel reports whether a model's table should exist given installed modules.
// Declaring module "" (unknown) is treated as kernel and always materialized.
func ShouldMaterializeModel(modelName string, installed map[string]struct{}) bool {
	mod := DeclaringModule(modelName)
	if mod == "" {
		return true
	}
	_, ok := installed[mod]
	return ok
}

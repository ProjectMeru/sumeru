package importgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/module"
)

func buildModelCatalog(repoRoot string, discovered map[string]*module.Addon) (map[string]string, error) {
	catalog := map[string]string{}

	scanDirs := []struct {
		dir    string
		module string
	}{
		{filepath.Join(repoRoot, "core", "orm"), "base"},
		{filepath.Join(repoRoot, "addons", "base", "models"), "base"},
	}

	names := make([]string, 0, len(discovered))
	for name := range discovered {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		addon := discovered[name]
		for _, sub := range []string{"models", "wizard"} {
			scanDirs = append(scanDirs, struct {
				dir    string
				module string
			}{filepath.Join(addon.Path, sub), name})
		}
	}

	for _, item := range scanDirs {
		if st, err := os.Stat(item.dir); err != nil || !st.IsDir() {
			continue
		}
		models, err := parseDirModels(item.dir)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", item.dir, err)
		}
		for _, m := range models {
			if m.ModelName == "" || m.ModelName == "-" || m.Extend {
				continue
			}
			if existing, ok := catalog[m.ModelName]; ok && existing != item.module {
				return nil, fmt.Errorf("model %q declared by both %q and %q", m.ModelName, existing, item.module)
			}
			catalog[m.ModelName] = item.module
		}
	}
	return catalog, nil
}

func validateInheritDepends(discovered map[string]*module.Addon, catalog map[string]string) error {
	names := make([]string, 0, len(discovered))
	for name := range discovered {
		names = append(names, name)
	}
	sort.Strings(names)

	var errs []string
	for _, name := range names {
		addon := discovered[name]
		if isBuiltinAddonPath(addon.Path) {
			continue
		}
		dependsSet := manifestDependsSet(&addon.Manifest)
		for _, sub := range []string{"models", "wizard"} {
			dir := filepath.Join(addon.Path, sub)
			if st, err := os.Stat(dir); err != nil || !st.IsDir() {
				continue
			}
			models, err := parseDirModels(dir)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			for _, m := range models {
				if !m.Extend || m.ModelName == "" || m.ModelName == "-" {
					continue
				}
				owner, ok := catalog[m.ModelName]
				if !ok {
					errs = append(errs, fmt.Sprintf("%s: inherit=%s: declaring module unknown (add model= in dependency or catalog)", name, m.ModelName))
					continue
				}
				if owner == name {
					continue
				}
				if _, ok := dependsSet[owner]; !ok {
					errs = append(errs, fmt.Sprintf("%s: inherit=%s requires manifest depends %q", name, m.ModelName, owner))
				}
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("inherit validation failed:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

func manifestDependsSet(m *module.Manifest) map[string]struct{} {
	out := map[string]struct{}{}
	for _, dep := range m.Depends {
		dep = strings.TrimSpace(dep)
		if dep != "" {
			out[dep] = struct{}{}
		}
	}
	return out
}

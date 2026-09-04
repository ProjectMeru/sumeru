package importgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/module"
)

func resolveAddonRoot(workspace, sumeruRoot, addonsRoot, name string) (string, error) {
	candidates := []string{
		filepath.Join(workspace, "addons", name),
		filepath.Join(sumeruRoot, "addons", name),
		filepath.Join(addonsRoot, name),
	}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, "manifest.json")); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("addon %q not found", name)
}

func collectDepends(workspace, sumeruRoot, addonsRoot, addonName string, seen map[string]struct{}) ([]string, error) {
	if _, ok := seen[addonName]; ok {
		return nil, nil
	}
	seen[addonName] = struct{}{}

	root, err := resolveAddonRoot(workspace, sumeruRoot, addonsRoot, addonName)
	if err != nil {
		return nil, err
	}
	m, err := module.ReadManifest(filepath.Join(root, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("manifest %s: %w", addonName, err)
	}

	var out []string
	for _, dep := range m.Depends {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		sub, err := collectDepends(workspace, sumeruRoot, addonsRoot, dep, seen)
		if err != nil {
			return nil, err
		}
		out = append(out, sub...)
		out = append(out, dep)
	}
	return out, nil
}

func dirsForModule(workspace, sumeruRoot, addonsRoot, moduleName string) []string {
	var dirs []string
	addDir := func(path string) {
		if st, err := os.Stat(path); err == nil && st.IsDir() {
			dirs = append(dirs, path)
		}
	}
	if moduleName == "base" {
		addDir(filepath.Join(sumeruRoot, "addons", "base", "models"))
		addDir(filepath.Join(sumeruRoot, "core", "orm"))
		return dirs
	}
	addDir(filepath.Join(sumeruRoot, "addons", moduleName, "models"))
	addDir(filepath.Join(sumeruRoot, "addons", moduleName, "wizard"))
	addDir(filepath.Join(addonsRoot, moduleName, "models"))
	addDir(filepath.Join(addonsRoot, moduleName, "wizard"))
	addDir(filepath.Join(workspace, "addons", moduleName, "models"))
	addDir(filepath.Join(workspace, "addons", moduleName, "wizard"))
	return dirs
}

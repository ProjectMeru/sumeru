// Package importgen implements sumeru-import-gen logic so it can be tested from sumeru/test.
package importgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/module"
	"sumeru/core/server/config"
)

// RunGen loads config, discovers addons, and writes the import file to dest.
func RunGen(repoRoot, configPath, outPath, packageName string) (written string, err error) {
	result, err := loadAndDiscover(repoRoot, configPath, outPath, packageName)
	if err != nil {
		return "", err
	}
	catalog, err := buildModelCatalog(repoRoot, result.discovered)
	if err != nil {
		return "", err
	}
	if err := validateInheritDepends(result.discovered, catalog); err != nil {
		return "", err
	}
	workspace := inferWorkspaceFromDiscovered(result.discovered)
	if workspace == "" {
		workspace = repoRoot
	}
	if err := validateRelationDepends(workspace, repoRoot, inferAddonsRoot(repoRoot), result.discovered, catalog); err != nil {
		return "", err
	}
	if err := module.ValidateDiscoveredAddons(result.discovered); err != nil {
		return "", fmt.Errorf("convention: %w", err)
	}
	if err := writeImportFile(result.dest, packageName, result.discovered); err != nil {
		return "", err
	}

	ormDir := filepath.Join(repoRoot, "core", "orm")
	ormmodelsDir := filepath.Join(repoRoot, "core", "ormmodels")
	if err := writeORMModelsRegistration(ormDir, ormmodelsDir, "base"); err != nil {
		return "", err
	}
	if err := writeORMZRefs(ormDir); err != nil {
		return "", err
	}

	addonsRoot := inferAddonsRoot(repoRoot)
	for _, addon := range result.discovered {
		if isBuiltinAddonPath(addon.Path) {
			continue
		}
		if err := generateRepoAddon(repoRoot, addonsRoot, result.discovered, addon); err != nil {
			return "", fmt.Errorf("%s: %w", addon.Manifest.Name, err)
		}
	}

	return result.dest, nil
}

func generateRepoAddon(repoRoot, addonsRoot string, discovered map[string]*module.Addon, addon *module.Addon) error {
	if err := writeAddonInit(discovered, addon); err != nil {
		return err
	}
	if err := writeAddonModelFiles(addon.Path, addon.Manifest.Name); err != nil {
		return err
	}
	modelsDir := filepath.Join(addon.Path, "models")
	return writeAddonZRefs(repoRoot, repoRoot, addonsRoot, addon.Manifest.Name, modelsDir, &addon.Manifest)
}

func inferAddonsRoot(sumeruRoot string) string {
	coreAddons := filepath.Clean(filepath.Join(sumeruRoot, "addons"))
	for _, p := range config.AppConfig.AddonPaths {
		abs, err := filepath.Abs(strings.TrimSpace(p))
		if err != nil || abs == coreAddons {
			continue
		}
		if _, err := os.Stat(filepath.Join(abs, "hr", "manifest.json")); err == nil {
			return abs
		}
	}
	return ""
}

func isBuiltinAddonPath(addonPath string) bool {
	return strings.Contains(filepath.Clean(addonPath), "module"+string(filepath.Separator)+"builtin")
}

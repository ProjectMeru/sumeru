package importgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/module"
)

func generateWorkspaceAddon(workspaceRoot, sumeruRoot, addonsRoot, addonName string, discovered map[string]*module.Addon) error {
	addon, ok := discovered[addonName]
	if !ok {
		return fmt.Errorf("addon %q not in discovery map", addonName)
	}
	addonRoot := addon.Path
	modelsDir := filepath.Join(addonRoot, "models")
	m := &addon.Manifest

	if isBuiltinAddonPath(addonRoot) {
		return nil
	}
	if err := writeAddonInit(discovered, addon); err != nil {
		return err
	}
	if err := writeAddonModelFiles(addonRoot, m.Name); err != nil {
		return err
	}
	return writeAddonZRefs(workspaceRoot, sumeruRoot, addonsRoot, addonName, modelsDir, m)
}

func writeAddonZRefs(workspaceRoot, sumeruRoot, addonsRoot, addonName, modelsDir string, m *module.Manifest) error {
	used, err := scanUsedRelationTypes(modelsDir)
	if err != nil {
		return err
	}

	zrefsPath := filepath.Join(modelsDir, "zrefs.go")
	if len(used) == 0 {
		_ = os.Remove(zrefsPath)
		return nil
	}

	seen := map[string]struct{}{}
	modules, err := collectDepends(workspaceRoot, sumeruRoot, addonsRoot, addonName, seen)
	if err != nil {
		return err
	}
	moduleSet := map[string]struct{}{}
	moduleSet[addonName] = struct{}{}
	for _, mod := range modules {
		moduleSet[mod] = struct{}{}
	}
	for _, dep := range m.Depends {
		moduleSet[strings.TrimSpace(dep)] = struct{}{}
	}
	flat := make([]string, 0, len(moduleSet))
	for mod := range moduleSet {
		if mod != "" {
			flat = append(flat, mod)
		}
	}

	rawRefs, err := collectModelRefs(workspaceRoot, sumeruRoot, addonsRoot, flat)
	if err != nil {
		return err
	}
	modelsImport, err := module.PackageImportPath(modelsDir)
	if err != nil {
		return err
	}
	var externalRefs []modelRef
	for _, ref := range rawRefs {
		if ref.ImportPath == modelsImport {
			continue
		}
		externalRefs = append(externalRefs, ref)
	}
	filtered := filterRefsForUsage(externalRefs, used)
	exported := buildExportedRefs(filtered, used)
	body := renderZRefs(exported)
	if body == "" {
		_ = os.Remove(zrefsPath)
		return nil
	}
	return os.WriteFile(zrefsPath, []byte(body), 0o644)
}

// RunWorkspaceGen writes imports, init.go, zmodels.go, and zrefs.go only under workspaceRoot.
// sumeruRoot and addonsRoot are read-only sources for cross-module relation type discovery.
func RunWorkspaceGen(workspaceRoot, sumeruRoot, addonsRoot, configPath, outPath, packageName string) (written string, err error) {
	workspaceRoot, err = filepath.Abs(workspaceRoot)
	if err != nil {
		return "", err
	}
	sumeruRoot, err = filepath.Abs(sumeruRoot)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(addonsRoot) != "" {
		addonsRoot, err = filepath.Abs(addonsRoot)
		if err != nil {
			return "", err
		}
	}

	result, err := loadAndDiscover(workspaceRoot, configPath, outPath, packageName)
	if err != nil {
		return "", err
	}

	workspaceModule, err := module.ReadModulePath(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("module name not found in %s/go.mod: %w", workspaceRoot, err)
	}
	_ = workspaceModule

	addonNames := make([]string, 0, len(result.discovered))
	for name := range result.discovered {
		addonNames = append(addonNames, name)
	}
	sort.Strings(addonNames)
	for _, name := range addonNames {
		if err := generateWorkspaceAddon(workspaceRoot, sumeruRoot, addonsRoot, name, result.discovered); err != nil {
			return "", fmt.Errorf("%s: %w", name, err)
		}
	}

	catalog, err := buildModelCatalog(sumeruRoot, result.discovered)
	if err != nil {
		return "", err
	}
	if err := validateInheritDepends(result.discovered, catalog); err != nil {
		return "", err
	}
	if err := validateRelationDepends(workspaceRoot, sumeruRoot, addonsRoot, result.discovered, catalog); err != nil {
		return "", err
	}
	if err := module.ValidateDiscoveredAddons(result.discovered); err != nil {
		return "", fmt.Errorf("convention: %w", err)
	}
	if err := writeImportFile(result.dest, packageName, result.discovered); err != nil {
		return "", err
	}

	return result.dest, nil
}

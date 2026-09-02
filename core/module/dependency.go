package module

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"sumeru/core/orm"
)

// SortDiscoveredAddonsTopo returns discovered addons in manifest dependency order.
func SortDiscoveredAddonsTopo(discovered map[string]*Addon) ([]*Addon, error) {
	return sortAddonsTopo(discovered)
}

// ModuleNamesTopo returns manifest technical names in dependency order.
func ModuleNamesTopo(discovered map[string]*Addon) ([]string, error) {
	topo, err := SortDiscoveredAddonsTopo(discovered)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(topo))
	for _, addon := range topo {
		names = append(names, addon.Manifest.Name)
	}
	return names, nil
}

func sortAddonsTopo(discovered map[string]*Addon) ([]*Addon, error) {
	remainingAddons := make(map[string]*Addon)
	for addonName, addon := range discovered {
		remainingAddons[addonName] = addon
	}
	var sortedAddons []*Addon
	for len(remainingAddons) > 0 {
		var addonCandidates []string
		for name := range remainingAddons {
			addon := remainingAddons[name]
			isSatisfied := true
			for _, dependencyName := range addon.Manifest.Depends {
				dependencyName = strings.TrimSpace(dependencyName)
				if dependencyName == "" || dependencyName == name {
					continue
				}
				if _, has := discovered[dependencyName]; !has {
					continue
				}
				if !containsAddonName(sortedAddons, dependencyName) {
					isSatisfied = false
					break
				}
			}
			if isSatisfied {
				addonCandidates = append(addonCandidates, name)
			}
		}
		if len(addonCandidates) == 0 {
			keys := make([]string, 0, len(remainingAddons))
			for k := range remainingAddons {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			addonCandidates = keys[:1]
		}
		sort.Strings(addonCandidates)
		selectedAddonName := addonCandidates[0]
		sortedAddons = append(sortedAddons, remainingAddons[selectedAddonName])
		delete(remainingAddons, selectedAddonName)
	}
	return sortedAddons, nil
}

func containsAddonName(addonList []*Addon, addonName string) bool {
	for _, addon := range addonList {
		if addon.Manifest.Name == addonName {
			return true
		}
	}
	return false
}

// ValidateDiscoveredDepends ensures every manifest depends entry exists on addons_path.
func ValidateDiscoveredDepends(discovered map[string]*Addon) error {
	if len(discovered) == 0 {
		return nil
	}
	names := make([]string, 0, len(discovered))
	for name := range discovered {
		names = append(names, name)
	}
	sort.Strings(names)

	var errs []string
	for _, name := range names {
		addon := discovered[name]
		for _, depName := range addon.Manifest.Depends {
			depName = strings.TrimSpace(depName)
			if depName == "" || depName == name {
				continue
			}
			if _, ok := discovered[depName]; !ok {
				errs = append(errs, fmt.Sprintf("%s: manifest depends on %q which is not on addons_path", name, depName))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("manifest depends validation failed:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

// ResolveInstallClosure returns moduleName and all transitive manifest depends in topological install order.
func ResolveInstallClosure(discovered map[string]*Addon, moduleName string) ([]string, error) {
	needed, err := installClosureSet(discovered, moduleName)
	if err != nil {
		return nil, err
	}
	topo, err := sortAddonsTopo(discovered)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(needed))
	for _, addon := range topo {
		if _, ok := needed[addon.Manifest.Name]; ok {
			out = append(out, addon.Manifest.Name)
		}
	}
	return out, nil
}

func installClosureSet(discovered map[string]*Addon, moduleName string) (map[string]struct{}, error) {
	if moduleName == "" {
		return nil, fmt.Errorf("module name required")
	}
	if _, ok := discovered[moduleName]; !ok {
		return nil, fmt.Errorf("unknown module %q", moduleName)
	}
	needed := map[string]struct{}{}
	var walk func(string) error
	walk = func(name string) error {
		addon, ok := discovered[name]
		if !ok {
			return fmt.Errorf("unknown module %q", name)
		}
		for _, depName := range addon.Manifest.Depends {
			depName = strings.TrimSpace(depName)
			if depName == "" || depName == name {
				continue
			}
			if _, has := discovered[depName]; !has {
				return fmt.Errorf("module %q depends on %q which is not on addons_path", name, depName)
			}
			if _, seen := needed[depName]; !seen {
				if err := walk(depName); err != nil {
					return err
				}
			}
		}
		needed[name] = struct{}{}
		return nil
	}
	if err := walk(moduleName); err != nil {
		return nil, err
	}
	return needed, nil
}

// OrderedManifestDepends returns direct manifest depends of moduleName in topological order.
func OrderedManifestDepends(discovered map[string]*Addon, moduleName string) []string {
	addon, ok := discovered[moduleName]
	if !ok {
		return nil
	}
	needed := map[string]struct{}{}
	for _, depName := range addon.Manifest.Depends {
		depName = strings.TrimSpace(depName)
		if depName == "" || depName == moduleName {
			continue
		}
		if _, has := discovered[depName]; !has {
			continue
		}
		needed[depName] = struct{}{}
	}
	if len(needed) == 0 {
		return nil
	}

	topo, err := sortAddonsTopo(discovered)
	if err != nil || len(topo) == 0 {
		names := make([]string, 0, len(needed))
		for n := range needed {
			names = append(names, n)
		}
		sort.Strings(names)
		return names
	}
	var out []string
	for _, dep := range topo {
		if _, ok := needed[dep.Manifest.Name]; ok {
			out = append(out, dep.Manifest.Name)
		}
	}
	return out
}

// ExpectedInitDependsImportPaths returns blank-import paths for manifest direct depends only.
func ExpectedInitDependsImportPaths(discovered map[string]*Addon, addon *Addon) ([]string, error) {
	var paths []string
	for _, depName := range OrderedManifestDepends(discovered, addon.Manifest.Name) {
		dep, ok := discovered[depName]
		if !ok {
			continue
		}
		importPath, err := PackageImportPath(dep.Path)
		if err != nil {
			return nil, fmt.Errorf("depends %q: %w", depName, err)
		}
		paths = append(paths, importPath)
	}
	return paths, nil
}

// missingInstalledDependencies lists manifest depends that are not installed in sys.module
// (or not registered). On-disk deps missing from DiscoveredAddons are ignored.
func missingInstalledDependencies(ctx context.Context, moduleName string) ([]string, error) {
	addon, ok := DiscoveredAddons[moduleName]
	if !ok {
		return nil, nil
	}
	var missingDependencies []string
	for _, dependencyName := range addon.Manifest.Depends {
		dependencyName = strings.TrimSpace(dependencyName)
		if dependencyName == "" || dependencyName == moduleName {
			continue
		}
		if _, has := DiscoveredAddons[dependencyName]; !has {
			continue
		}
		moduleRow, err := moduleRow(ctx, dependencyName)
		if err != nil {
			if err == sql.ErrNoRows {
				missingDependencies = append(missingDependencies, dependencyName)
				continue
			}
			return nil, err
		}
		if moduleStateString(moduleRow) != "installed" {
			missingDependencies = append(missingDependencies, dependencyName)
		}
	}
	return missingDependencies, nil
}

func installedModuleDependingOn(ctx context.Context, targetModuleName string) (string, error) {
	moduleRows, err := orm.DB.QueryContext(ctx,
		`SELECT name, state FROM `+orm.MustQuotedTableName("sys.module")+` WHERE state = 'installed' AND name <> $1`,
		targetModuleName,
	)
	if err != nil {
		return "", err
	}
	defer moduleRows.Close()

	for moduleRows.Next() {
		var moduleName, state string
		if err := moduleRows.Scan(&moduleName, &state); err != nil {
			return "", err
		}
		addon, ok := DiscoveredAddons[moduleName]
		if !ok {
			continue
		}
		for _, dependencyName := range addon.Manifest.Depends {
			if strings.TrimSpace(dependencyName) == targetModuleName {
				return moduleName, nil
			}
		}
	}
	return "", moduleRows.Err()
}

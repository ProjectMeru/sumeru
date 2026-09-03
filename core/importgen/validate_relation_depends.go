package importgen

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/module"
)

func validateRelationDepends(workspace, sumeruRoot, addonsRoot string, discovered map[string]*module.Addon, catalog map[string]string) error {
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
		modelsDir := filepath.Join(addon.Path, "models")
		used, err := scanUsedRelationTypes(modelsDir)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if len(used) == 0 {
			continue
		}

		closure, err := module.ResolveInstallClosure(discovered, name)
		if err != nil {
			return err
		}
		dependsSet := map[string]struct{}{}
		for _, mod := range closure {
			if mod != name {
				dependsSet[mod] = struct{}{}
			}
		}

		modules, err := collectDepends(workspace, sumeruRoot, addonsRoot, name, map[string]struct{}{})
		if err != nil {
			return err
		}
		moduleSet := map[string]struct{}{name: {}}
		for _, mod := range modules {
			moduleSet[mod] = struct{}{}
		}
		for _, dep := range addon.Manifest.Depends {
			moduleSet[strings.TrimSpace(dep)] = struct{}{}
		}
		flat := make([]string, 0, len(moduleSet))
		for mod := range moduleSet {
			if mod != "" {
				flat = append(flat, mod)
			}
		}

		rawRefs, err := collectModelRefs(workspace, sumeruRoot, addonsRoot, flat)
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
		for _, ref := range filtered {
			owner := catalog[ref.TechnicalModel]
			if owner == "" || owner == name {
				continue
			}
			if _, ok := dependsSet[owner]; !ok {
				errs = append(errs, fmt.Sprintf("%s: relation %s (%s) requires manifest depends %q", name, ref.TechnicalModel, ref.GoName, owner))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("relation depends validation failed:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

func inferWorkspaceFromDiscovered(discovered map[string]*module.Addon) string {
	for _, addon := range discovered {
		root, err := module.FindRepoRoot(addon.Path)
		if err != nil {
			continue
		}
		mod, err := module.ReadModulePath(root)
		if err != nil || mod == "" || mod == "sumeru" {
			continue
		}
		return root
	}
	return ""
}

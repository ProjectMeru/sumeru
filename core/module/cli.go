package module

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

// UpdateModuleData reloads XML for an installed module (-u): drops module-linked UI metadata then re-syncs.
func UpdateModuleData(ctx context.Context, name string) error {
	installMu.Lock()
	defer installMu.Unlock()

	if _, ok := DiscoveredAddons[name]; !ok {
		return fmt.Errorf("unknown module %q", name)
	}
	row, err := moduleRow(ctx, name)
	if err != nil {
		return err
	}
	state := moduleStateString(row)
	if state != "installed" && state != "to_upgrade" {
		return fmt.Errorf("module %q is not installed", name)
	}
	missing, err := missingInstalledDependencies(ctx, name)
	if err != nil {
		return fmt.Errorf("module %q: %w", name, err)
	}
	if len(missing) > 0 {
		// e.g. -u all while a leaf is marked installed without its deps — skip reload (no DB churn).
		return nil
	}
	if err := reloadInstalledDependencies(ctx, name); err != nil {
		return err
	}
	return reloadModuleData(ctx, name, moduleReloadUpdate)
}

// RunModuleCLI runs -i (install) and -u (update) lists. Install runs before update.
// For -u, the token "all" (case-insensitive) means every installed (or to_upgrade) module
// present on disk, in dependency order. Uninstalled modules are never updated.
func RunModuleCLI(installCSV, updateCSV string) error {
	ctx := orm.AuditedBypass(context.Background(), "module.cli")

	installNames, err := expandInstallModuleNames(ctx, splitCSV(installCSV))
	if err != nil {
		return err
	}

	for _, name := range installNames {
		if err := InstallModuleByName(ctx, name); err != nil {
			return fmt.Errorf("install %q: %w", name, err)
		}
	}
	updateList, err := expandUpdateModuleNames(ctx, splitCSV(updateCSV))
	if err != nil {
		return err
	}
	if len(updateList) > 0 {
		applog.InfoMsg(ctx, "module", "update",
			fmt.Sprintf("Updating %d installed module(s)", len(updateList)),
			map[string]interface{}{"modules": strings.Join(updateList, ", ")})
	}
	for _, name := range updateList {
		if err := UpdateModuleData(ctx, name); err != nil {
			return fmt.Errorf("update %q: %w", name, err)
		}
	}
	return nil
}

func listInstalledModuleNames(ctx context.Context) ([]string, error) {
	q := `SELECT name FROM ` + orm.MustQuotedTableName("sys.module") +
		` WHERE state IN ('installed', 'to_upgrade') ORDER BY name`
	rows, err := orm.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list installed modules: %w", err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

// expandUpdateModuleNames resolves -u all into installed module names, dedupes, and orders by addon dependency topo.
// Installed rows not present on disk (stale sys.module) are skipped for "all" only.
func expandUpdateModuleNames(ctx context.Context, parts []string) ([]string, error) {
	if len(parts) == 0 {
		return nil, nil
	}
	var installedCache []string
	var installedErr error
	fetchInstalled := func() ([]string, error) {
		if installedErr != nil {
			return nil, installedErr
		}
		if installedCache == nil {
			installedCache, installedErr = listInstalledModuleNames(ctx)
		}
		return installedCache, installedErr
	}

	seen := make(map[string]struct{})
	var raw []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.EqualFold(p, "all") {
			names, err := fetchInstalled()
			if err != nil {
				return nil, err
			}
			for _, n := range names {
				if _, known := DiscoveredAddons[n]; !known {
					continue
				}
				if _, ok := seen[n]; ok {
					continue
				}
				seen[n] = struct{}{}
				raw = append(raw, n)
			}
			continue
		}
		// Explicitly named module in -u
		if _, ok := seen[p]; ok {
			continue
		}
		if _, onDisk := DiscoveredAddons[p]; !onDisk {
			syncWarn(ctx, "Warning: module %q is installed in DB but not on disk — skipped for -u", p)
			continue
		}
		// Check if it's actually installed
		installed, err := fetchInstalled()
		if err != nil {
			return nil, err
		}
		isInstalled := false
		for _, inst := range installed {
			if inst == p {
				isInstalled = true
				break
			}
		}
		if !isInstalled {
			// User requested: if module is not installed then no need to install (and no error)
			continue
		}

		seen[p] = struct{}{}
		raw = append(raw, p)
	}
	return orderModulesForUpdate(raw), nil
}

func orderModulesForUpdate(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	topo, err := sortAddonsTopo(DiscoveredAddons)
	if err != nil || len(topo) == 0 {
		out := append([]string(nil), names...)
		sort.Strings(out)
		return out
	}
	var ordered []string
	for _, a := range topo {
		n := a.Manifest.Name
		if _, ok := set[n]; !ok {
			continue
		}
		ordered = append(ordered, n)
		delete(set, n)
	}
	rest := make([]string, 0, len(set))
	for n := range set {
		rest = append(rest, n)
	}
	sort.Strings(rest)
	return append(ordered, rest...)
}

// expandInstallModuleNames resolves -i all into uninstalled module names in dependency order.
func expandInstallModuleNames(ctx context.Context, parts []string) ([]string, error) {
	if len(parts) == 0 {
		return nil, nil
	}
	var installedSet map[string]struct{}
	loadInstalled := func() error {
		if installedSet != nil {
			return nil
		}
		installed, err := listInstalledModuleNames(ctx)
		if err != nil {
			return err
		}
		installedSet = make(map[string]struct{}, len(installed))
		for _, n := range installed {
			installedSet[n] = struct{}{}
		}
		return nil
	}

	seen := make(map[string]struct{})
	var raw []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.EqualFold(p, "all") {
			if err := loadInstalled(); err != nil {
				return nil, err
			}
			topo, err := sortAddonsTopo(DiscoveredAddons)
			if err != nil {
				return nil, err
			}
			for _, a := range topo {
				n := a.Manifest.Name
				if _, ok := installedSet[n]; ok {
					continue
				}
				if _, ok := seen[n]; ok {
					continue
				}
				seen[n] = struct{}{}
				raw = append(raw, n)
			}
			continue
		}
		if _, ok := DiscoveredAddons[p]; !ok {
			return nil, fmt.Errorf("unknown module %q", p)
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		raw = append(raw, p)
	}
	return orderModulesForUpdate(raw), nil
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

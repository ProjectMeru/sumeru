package module

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func runAutoInstallPass(ctx context.Context) error {
	for {
		changed := false
		topo, err := SortDiscoveredAddonsTopo(DiscoveredAddons)
		if err != nil {
			return err
		}
		for _, addon := range topo {
			if !addon.Manifest.IsAutoInstall() {
				continue
			}
			name := addon.Manifest.Name
			moduleRow, err := moduleRow(ctx, name)
			if err != nil {
				if err == sql.ErrNoRows {
					continue
				}
				return err
			}
			if moduleStateString(moduleRow) == "installed" {
				continue
			}
			missing, err := missingInstalledDependencies(ctx, name)
			if err != nil {
				return err
			}
			if len(missing) > 0 {
				continue
			}
			if err := installModuleUnlocked(ctx, name); err != nil {
				return fmt.Errorf("auto-install %q: %w", name, err)
			}
			changed = true
		}
		if !changed {
			return nil
		}
	}
}

// AllDependsInstalled reports whether every manifest depend of moduleName is installed.
func AllDependsInstalled(ctx context.Context, moduleName string) (bool, error) {
	missing, err := missingInstalledDependencies(ctx, moduleName)
	if err != nil {
		return false, err
	}
	return len(missing) == 0, nil
}

// AutoInstallCandidateNames returns uninstalled auto_install modules whose depends are satisfied.
func AutoInstallCandidateNames(ctx context.Context) ([]string, error) {
	topo, err := SortDiscoveredAddonsTopo(DiscoveredAddons)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, addon := range topo {
		if !addon.Manifest.IsAutoInstall() {
			continue
		}
		name := strings.TrimSpace(addon.Manifest.Name)
		moduleRow, err := moduleRow(ctx, name)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
		if moduleStateString(moduleRow) == "installed" {
			continue
		}
		ok, err := AllDependsInstalled(ctx, name)
		if err != nil {
			return nil, err
		}
		if ok {
			names = append(names, name)
		}
	}
	return names, nil
}

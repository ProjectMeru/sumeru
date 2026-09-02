package module

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"sumeru/addons/mail"
	"sumeru/core/applog"
	"sumeru/core/metrics"
	"sumeru/core/orm"
)

// Install pipeline per module: resolveDeps → markPending → syncSchema → loadData → finalize.
// XML record sync (views, menus, actions) runs inside loadData via SyncToDB.

// InstallModuleByName marks the module installed and loads its XML (and dependencies first).
func InstallModuleByName(ctx context.Context, moduleName string) error {
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), "sys.module", "write"); err != nil {
		return err
	}
	systemContext := orm.ContextWithBypass(ctx, true)
	installMu.Lock()
	defer installMu.Unlock()
	if closure, err := ResolveInstallClosure(DiscoveredAddons, moduleName); err != nil {
		return err
	} else {
		var deps []string
		for _, name := range closure {
			if name != moduleName {
				deps = append(deps, name)
			}
		}
		if len(deps) > 0 {
			applog.InfoMsg(systemContext, "module", "install",
				fmt.Sprintf("Installing %s (+ deps: %s)", moduleName, strings.Join(deps, ", ")),
				map[string]interface{}{"module": moduleName, "depends": deps})
		} else {
			applog.InfoMsg(systemContext, "module", "install",
				fmt.Sprintf("Installing %s", moduleName),
				map[string]interface{}{"module": moduleName})
		}
	}
	if err := installModuleUnlocked(systemContext, moduleName); err != nil {
		return err
	}
	return runAutoInstallPass(systemContext)
}

func installModuleUnlocked(ctx context.Context, moduleName string) error {
	start := time.Now()
	defer func() {
		metrics.ObserveDuration("sumeru_module_install_duration_seconds", time.Since(start))
	}()
	if moduleName == "" {
		return fmt.Errorf("module name required")
	}
	addon, ok := DiscoveredAddons[moduleName]
	if !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}

	for _, dependencyName := range addon.Manifest.Depends {
		dependencyName = strings.TrimSpace(dependencyName)
		if dependencyName == "" || dependencyName == addon.Manifest.Name {
			continue
		}
		if _, has := DiscoveredAddons[dependencyName]; !has {
			return fmt.Errorf("module %q depends on %q which is not on addons_path", moduleName, dependencyName)
		}
		moduleRow, err := moduleRow(ctx, dependencyName)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("dependency %q is not registered", dependencyName)
			}
			return err
		}
		if moduleStateString(moduleRow) != "installed" {
			if err := installModuleUnlocked(ctx, dependencyName); err != nil {
				return fmt.Errorf("install dependency %q: %w", dependencyName, err)
			}
		}
	}

	if err := reloadInstalledDependencies(ctx, moduleName); err != nil {
		return err
	}

	return reloadModuleData(ctx, moduleName, moduleReloadInstall)
}

type moduleReloadMode int

const (
	moduleReloadInstall moduleReloadMode = iota
	moduleReloadUpdate
)

// reloadModuleData syncs schema and XML for install (-i) or update (-u).
func reloadModuleData(ctx context.Context, moduleName string, mode moduleReloadMode) error {
	addon, ok := DiscoveredAddons[moduleName]
	if !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}

	if err := markModuleReloadPending(ctx, moduleName, mode); err != nil {
		return err
	}
	if err := syncModuleSchema(ctx, moduleName, mode); err != nil {
		return err
	}
	if err := loadModuleXMLData(ctx, moduleName, mode, addon); err != nil {
		return err
	}
	return finalizeModuleReload(ctx, moduleName, mode)
}

func markModuleReloadPending(ctx context.Context, moduleName string, mode moduleReloadMode) error {
	switch mode {
	case moduleReloadInstall:
		return setModuleState(ctx, moduleName, "to_install", true)
	case moduleReloadUpdate:
		return setModuleStateOnly(ctx, moduleName, "to_upgrade")
	default:
		return fmt.Errorf("unknown reload mode")
	}
}

func syncModuleSchema(ctx context.Context, moduleName string, mode moduleReloadMode) error {
	if err := orm.SyncRegistrySchemaForModule(moduleName); err != nil {
		_ = setModuleLastError(ctx, moduleName, err.Error())
		if mode == moduleReloadInstall {
			return FatalSync(moduleName, "schema sync", err)
		}
		return fmt.Errorf("schema sync: %w", err)
	}
	return nil
}

func loadModuleXMLData(ctx context.Context, moduleName string, mode moduleReloadMode, addon *Addon) error {
	if mode == moduleReloadUpdate {
		if err := deleteModuleMetadata(ctx, moduleName); err != nil {
			return err
		}
	}
	ctx = ContextWithSyncMode(ctx, mode)
	return recordSyncToDBResult(ctx, moduleName, addon.SyncToDB(ctx))
}

func finalizeModuleReload(ctx context.Context, moduleName string, mode moduleReloadMode) error {
	switch mode {
	case moduleReloadInstall:
		if err := setModuleState(ctx, moduleName, "installed", true); err != nil {
			return err
		}
		orm.InvalidateRuleCache()
		mail.LogModuleEvent(ctx, moduleName, "Installed", "")
	case moduleReloadUpdate:
		if err := setModuleStateOnly(ctx, moduleName, "installed"); err != nil {
			return err
		}
		mail.LogModuleEvent(ctx, moduleName, "Updated", "module data reloaded")
	}
	return nil
}

// UninstallModuleByName removes XML-linked metadata for the module and marks it uninstalled.
func UninstallModuleByName(ctx context.Context, moduleName string) error {
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), "sys.module", "write"); err != nil {
		return err
	}
	systemContext := orm.ContextWithBypass(ctx, true)
	installMu.Lock()
	defer installMu.Unlock()

	if orm.IsPlatformModule(moduleName) {
		return fmt.Errorf("cannot uninstall platform module %q", moduleName)
	}
	if _, ok := DiscoveredAddons[moduleName]; !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}

	if dependency, err := installedModuleDependingOn(systemContext, moduleName); err != nil {
		return err
	} else if dependency != "" {
		return fmt.Errorf("module %q is required by installed module %q; uninstall that first", moduleName, dependency)
	}

	if err := setModuleToRemove(systemContext, moduleName); err != nil {
		return err
	}

	if err := deleteModuleMetadata(systemContext, moduleName); err != nil {
		return err
	}

	if err := setModuleState(systemContext, moduleName, "uninstalled", true); err != nil {
		return err
	}
	mail.LogModuleEvent(systemContext, moduleName, "Uninstalled", "")
	return nil
}

func deleteModuleMetadata(ctx context.Context, moduleName string) error {
	modelDataTable := orm.MustQuotedTableName("sys.model.data")

	viewTable, err := orm.QuotedTableName("sys.view")
	if err != nil {
		return fmt.Errorf("delete sys.view: %w", err)
	}
	if _, err := orm.DB.ExecContext(ctx, ModuleViewDeleteQuery(viewTable, modelDataTable), moduleName); err != nil {
		return fmt.Errorf("delete sys.view: %w", err)
	}

	modelNames := []string{"sys.menu", "sys.action.window", "sys.access", "sys.rule", "sys.approval.rule"}
	for _, modelName := range modelNames {
		tableName, err := orm.QuotedTableName(modelName)
		if err != nil {
			return fmt.Errorf("delete %s: %w", modelName, err)
		}
		deleteQuery := `DELETE FROM ` + tableName + ` WHERE id IN (SELECT core_id FROM ` + modelDataTable + ` WHERE module = $1 AND model = $2)`
		if _, err := orm.DB.ExecContext(ctx, deleteQuery, moduleName, modelName); err != nil {
			return fmt.Errorf("delete %s: %w", modelName, err)
		}
	}
	if _, err := orm.DB.ExecContext(ctx, `DELETE FROM `+modelDataTable+` WHERE module = $1`, moduleName); err != nil {
		return err
	}
	return nil
}

// ModuleViewDeleteQuery deletes sys.view rows owned solely by moduleName.
// Rows whose core_id is also linked from another module (view inherit extensions) are kept.
func ModuleViewDeleteQuery(viewTable, modelDataTable string) string {
	return `DELETE FROM ` + viewTable + ` WHERE id IN (
		SELECT md.core_id FROM ` + modelDataTable + ` md
		WHERE md.module = $1 AND md.model = 'sys.view'
		AND NOT EXISTS (
			SELECT 1 FROM ` + modelDataTable + ` other
			WHERE other.model = 'sys.view' AND other.core_id = md.core_id AND other.module <> $1
		)
	)`
}

// SetModuleActive toggles visibility of menus for an installed module without removing data.
func SetModuleActive(ctx context.Context, moduleName string, active bool) error {
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), "sys.module", "write"); err != nil {
		return err
	}
	systemContext := orm.ContextWithBypass(ctx, true)
	installMu.Lock()
	defer installMu.Unlock()

	if moduleName == KernelModule && !active {
		return fmt.Errorf("cannot deactivate core module %q", KernelModule)
	}
	if _, ok := DiscoveredAddons[moduleName]; !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}

	moduleRow, err := moduleRow(systemContext, moduleName)
	if err != nil {
		return err
	}
	if moduleStateString(moduleRow) != "installed" {
		return fmt.Errorf("module %q is not installed; activate/install it first", moduleName)
	}

	if err := setModuleActiveOnly(systemContext, moduleName, active); err != nil {
		return err
	}
	if active {
		mail.LogModuleEvent(systemContext, moduleName, "Activated", "")
	} else {
		mail.LogModuleEvent(systemContext, moduleName, "Deactivated", "")
	}
	return nil
}

// ListModules returns sys.module rows for the Apps UI (non-application modules included for completeness).
func ListModules(ctx context.Context) ([]map[string]interface{}, error) {
	moduleTable := orm.MustQuotedTableName("sys.module")
	categoryTable := orm.MustQuotedTableName("sys.module.category")
	moduleRows, err := orm.DB.QueryContext(ctx,
		`SELECT m.id, m.name, m.display_name, m.author, m.version, m.description, m.state, m.application, m.active, m.category_id, c.name AS category_name FROM `+
			moduleTable+` m LEFT JOIN `+categoryTable+` c ON c.id = m.category_id ORDER BY m.application DESC, m.name`,
	)
	if err != nil {
		return nil, err
	}
	defer moduleRows.Close()

	var moduleList []map[string]interface{}
	columnNames := []string{"id", "name", "display_name", "author", "version", "description", "state", "application", "active", "category_id", "category_name"}

	for moduleRows.Next() {
		var id int64
		var name, display, author, version, state string
		var desc sql.NullString
		var categoryName sql.NullString
		var categoryID sql.NullInt64
		var application, active bool
		if err := moduleRows.Scan(&id, &name, &display, &author, &version, &desc, &state, &application, &active, &categoryID, &categoryName); err != nil {
			return nil, err
		}
		moduleMap := make(map[string]interface{})
		moduleMap["id"] = id
		moduleMap[columnNames[1]] = name
		moduleMap[columnNames[2]] = display
		moduleMap[columnNames[3]] = author
		moduleMap[columnNames[4]] = version
		if desc.Valid {
			moduleMap[columnNames[5]] = desc.String
		} else {
			moduleMap[columnNames[5]] = ""
		}
		moduleMap[columnNames[6]] = state
		moduleMap[columnNames[7]] = application
		moduleMap[columnNames[8]] = active
		if categoryID.Valid {
			moduleMap[columnNames[9]] = categoryID.Int64
		}
		if categoryName.Valid {
			moduleMap[columnNames[10]] = categoryName.String
		} else {
			moduleMap[columnNames[10]] = ""
		}
		moduleList = append(moduleList, moduleMap)
	}
	return moduleList, moduleRows.Err()
}

// reloadInstalledDependencies re-syncs XML for installed transitive dependencies before
// installing or updating a module (e.g. ensure mail.activity views exist when CRM loads).
func reloadInstalledDependencies(ctx context.Context, moduleName string) error {
	for _, depName := range transitiveDependencies(moduleName) {
		moduleRow, err := moduleRow(ctx, depName)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return err
		}
		if moduleStateString(moduleRow) != "installed" {
			continue
		}
		if err := reloadModuleData(ctx, depName, moduleReloadUpdate); err != nil {
			return fmt.Errorf("update dependency %q: %w", depName, err)
		}
	}
	return nil
}

// transitiveDependencies returns manifest depends (recursive) in topological order.
func transitiveDependencies(moduleName string) []string {
	needed, err := installClosureSet(DiscoveredAddons, moduleName)
	if err != nil {
		return nil
	}
	delete(needed, moduleName)

	topo, err := sortAddonsTopo(DiscoveredAddons)
	if err != nil || len(topo) == 0 {
		names := make([]string, 0, len(needed))
		for n := range needed {
			names = append(names, n)
		}
		sort.Strings(names)
		return names
	}
	out := make([]string, 0, len(needed))
	for _, addon := range topo {
		if _, ok := needed[addon.Manifest.Name]; ok {
			out = append(out, addon.Manifest.Name)
		}
	}
	return out
}

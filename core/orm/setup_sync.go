package orm

import (
	"context"
	"fmt"
	"sort"
)

// InitialSetupModelNames is the subset synced during /setup/init; full registry sync is SyncRegistrySchema.
var InitialSetupModelNames = []string{
	"app.log",
	"core.company",
	"sys.module.category", // before core.group (category_id); kernel metadata
	"core.group",
	"core.partner",
	"core.user",
	"mail.message",
	"sys.access",
	"sys.action.window",
	"sys.action.url",
	"sys.approval.rule",
	"sys.field",
	"sys.menu",
	"sys.model",
	"sys.model.data",
	"sys.module",
	"sys.rule",
	"sys.session",
	"sys.view",
}

// SyncModelsInitialSetup creates tables only for InitialSetupModelNames (first-run /setup).
func SyncModelsInitialSetup() error {
	ctx := AuditedBypass(context.Background(), "setup.sync")
	for _, name := range InitialSetupModelNames {
		m, ok := Registry[name]
		if !ok {
			return fmt.Errorf("initial setup: model %q is not registered (build must include sumeru/addons/base)", name)
		}
		if err := createTable(ctx, m); err != nil {
			return fmt.Errorf("create table %s: %w", name, err)
		}
	}
	return nil
}

// SyncRegistrySchemaInitialSetup applies schema drift fixes only for InitialSetupModelNames.
func SyncRegistrySchemaInitialSetup() error {
	return SyncRegistrySchemaForNames(InitialSetupModelNames)
}

// SyncRegistrySchemaForNames runs schema sync for an explicit model list (order is sorted for stability).
func SyncRegistrySchemaForNames(modelNames []string) error {
	if DB == nil {
		return nil
	}
	ctx := AuditedBypass(context.Background(), "setup.sync")
	names := append([]string(nil), modelNames...)
	sort.Strings(names)
	for _, name := range names {
		m, ok := Registry[name]
		if !ok {
			return fmt.Errorf("schema sync: model %q not registered", name)
		}
		if err := syncModelSchema(ctx, m); err != nil {
			return fmt.Errorf("schema sync %s: %w", name, err)
		}
	}
	return ensureExtraIndexes(ctx)
}

// ModelsForModuleSchemaSync returns (names, true) when install should only touch those models;
// (nil, false) means sync the full registry (unknown modules with no registered models).
func ModelsForModuleSchemaSync(moduleName string) ([]string, bool) {
	owned := ModelsOwnedByModule(moduleName)
	extended := ModelsExtendedByModule(moduleName)
	if len(owned) == 0 && len(extended) == 0 {
		return nil, false
	}
	seen := make(map[string]struct{}, len(owned)+len(extended))
	var names []string
	for _, name := range append(owned, extended...) {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, true
}

// SyncRegistrySchemaForModule runs schema sync for models owned by the module, or the full registry.
func SyncRegistrySchemaForModule(moduleName string) error {
	names, scoped := ModelsForModuleSchemaSync(moduleName)
	if scoped {
		return SyncRegistrySchemaForNames(names)
	}
	return SyncRegistrySchema()
}

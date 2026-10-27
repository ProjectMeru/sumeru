package sdk

import (
	"context"

	"sumeru/core/orm"
)

// ModelRegistered reports whether name is in the model registry.
func ModelRegistered(name string) bool {
	_, ok := orm.Registry[name]
	return ok
}

// ModelFieldNames returns declared field names. ok is false when the model is missing.
func ModelFieldNames(modelName string) ([]string, bool) {
	inst := orm.RegistryModel(modelName)
	if inst == nil {
		return nil, false
	}
	names := make([]string, 0, len(inst.Fields()))
	for _, field := range inst.Fields() {
		if field.Name != "" {
			names = append(names, field.Name)
		}
	}
	return names, true
}

// InstalledModuleNames returns technical names of installed, active modules.
func InstalledModuleNames(ctx context.Context) (map[string]struct{}, error) {
	return orm.InstalledModuleNames(ctx)
}

// ShouldMaterializeModel reports whether a model's table should exist for the installed set.
func ShouldMaterializeModel(modelName string, installed map[string]struct{}) bool {
	return orm.ShouldMaterializeModel(modelName, installed)
}

// EnsureModelColumns adds missing physical columns for extra field definitions.
func EnsureModelColumns(ctx context.Context, modelName string, extra []orm.FieldDefinition) error {
	inst, ok := orm.Registry[modelName]
	if !ok || inst == nil {
		return nil
	}
	return orm.EnsureModelColumns(ctx, inst, extra)
}

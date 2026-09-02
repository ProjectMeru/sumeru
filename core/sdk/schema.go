package sdk

import (
	"context"

	"sumeru/core/orm"
)

// EnsureModelColumns adds missing physical columns for extra field definitions.
func EnsureModelColumns(ctx context.Context, modelName string, extra []orm.FieldDefinition) error {
	inst, ok := orm.Registry[modelName]
	if !ok || inst == nil {
		return nil
	}
	return orm.EnsureModelColumns(ctx, inst, extra)
}

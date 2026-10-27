package sdk

import (
	"context"

	"sumeru/core/orm"
)

// ReadGroupSpec defines grouping and aggregation.
type ReadGroupSpec = orm.ReadGroupSpec

// ReadGroupField is one aggregate column.
type ReadGroupField = orm.ReadGroupField

// ReadGroupRow is one grouped result row.
type ReadGroupRow = orm.ReadGroupRow

// ReadGroup aggregates records by group fields.
func ReadGroup(ctx context.Context, modelName string, domain [][]interface{}, spec ReadGroupSpec) ([]ReadGroupRow, error) {
	return orm.ReadGroup(ctx, modelName, domain, spec)
}

// RegisterReadGroupWrapper sets the optional ReadGroup wrapper (enterprise ops).
func RegisterReadGroupWrapper(w orm.ReadGroupWrapper) {
	orm.RegisterReadGroupWrapper(w)
}

// ClearReadGroupWrapper removes the wrapper (tests).
func ClearReadGroupWrapper() {
	orm.ClearReadGroupWrapper()
}

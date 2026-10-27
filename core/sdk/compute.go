package sdk

import (
	"context"

	"sumeru/core/orm"
)

// RegisterCompute registers a computed field. deps are the fields that trigger recomputation.
func RegisterCompute(model, field string, deps []string, fn func(ctx context.Context, rec map[string]interface{}) (interface{}, error)) {
	orm.RegisterCompute(model, field, deps, fn)
}

// ComputeContext carries the record id for compute helpers that need it.
// Request-scoped context is passed separately to compute handlers.
// This struct intentionally does not store context.Context.
type ComputeContext struct {
	ID int
}

// NewComputeContext builds a compute context for a record.
func NewComputeContext(id int) ComputeContext {
	return ComputeContext{ID: id}
}

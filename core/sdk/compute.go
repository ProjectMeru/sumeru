package sdk

// ComputeContext carries the record id for compute helpers that need it.
// Request-scoped context is passed separately to compute handlers (orm.ComputeFunc);
// this struct intentionally does not store context.Context.
type ComputeContext struct {
	ID int
}

// NewComputeContext builds a compute context for a record.
func NewComputeContext(id int) ComputeContext {
	return ComputeContext{ID: id}
}

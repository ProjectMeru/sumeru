package sdk

import (
	"context"

	"sumeru/core/orm"
)

// RegisterConstraint registers a cross-field check for create and write.
func RegisterConstraint(modelName, name string, fn func(ctx context.Context, rec map[string]interface{}) error) {
	orm.RegisterConstraint(modelName, name, fn)
}

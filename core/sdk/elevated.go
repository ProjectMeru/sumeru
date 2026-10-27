package sdk

import (
	"context"

	"sumeru/core/orm"
)

// WithElevated runs fn with an audited security bypass.
func WithElevated(ctx context.Context, reason string, fn func(context.Context) error) error {
	return orm.WithElevated(ctx, reason, fn)
}

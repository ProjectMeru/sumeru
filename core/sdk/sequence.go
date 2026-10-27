package sdk

import (
	"context"

	"sumeru/core/orm"
)

// NextSequence returns the next formatted number for code and advances the counter.
func NextSequence(ctx context.Context, code string) (string, error) {
	return orm.NextSequence(ctx, code)
}

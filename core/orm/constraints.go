package orm

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type ConstraintFunc func(ctx context.Context, rec map[string]interface{}) error

var (
	constraintMu sync.RWMutex
	constraints  = map[string][]ConstraintFunc{} // model -> handlers
)

// RegisterConstraint registers a cross-field validation handler for create/write.
func RegisterConstraint(modelName, name string, fn ConstraintFunc) {
	if fn == nil {
		return
	}
	modelName = normalizeModelName(modelName)
	constraintMu.Lock()
	defer constraintMu.Unlock()
	constraints[modelName] = append(constraints[modelName], fn)
}

func runConstraints(ctx context.Context, modelName string, rec map[string]interface{}) error {
	modelName = normalizeModelName(modelName)
	constraintMu.RLock()
	handlers := constraints[modelName]
	constraintMu.RUnlock()
	for _, fn := range handlers {
		if err := fn(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

func normalizeModelName(modelName string) string {
	return strings.TrimSpace(modelName)
}

// ValidateConstraints runs registered constraints; exported for tests.
func ValidateConstraints(ctx context.Context, modelName string, rec map[string]interface{}) error {
	if rec == nil {
		return fmt.Errorf("record required")
	}
	return runConstraints(ctx, modelName, rec)
}

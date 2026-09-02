package orm

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ComputeFunc derives a stored or read-time field value from a record map.
type ComputeFunc func(ctx context.Context, rec map[string]interface{}) (interface{}, error)

type computeSpec struct {
	deps []string
	fn   ComputeFunc
}

var (
	computeMu sync.RWMutex
	computes  = map[string]map[string]computeSpec{} // model -> field -> spec
)

// RegisterCompute registers a computed field for model.
// deps lists field names that trigger recomputation when present in writes.
func RegisterCompute(model, field string, deps []string, fn ComputeFunc) {
	model = strings.TrimSpace(model)
	field = strings.TrimSpace(field)
	if model == "" || field == "" || fn == nil {
		return
	}
	computeMu.Lock()
	defer computeMu.Unlock()
	if computes[model] == nil {
		computes[model] = map[string]computeSpec{}
	}
	computes[model][field] = computeSpec{deps: append([]string(nil), deps...), fn: fn}
}

// ApplyComputes fills registered computed fields on rec (in place).
func ApplyComputes(ctx context.Context, model string, rec map[string]interface{}) error {
	if rec == nil {
		return nil
	}
	specs := snapshotComputeSpecs(model)
	if len(specs) == 0 {
		return nil
	}
	for field, spec := range specs {
		val, err := spec.fn(ctx, rec)
		if err != nil {
			return fmt.Errorf("compute %s.%s: %w", model, field, err)
		}
		rec[field] = val
	}
	return nil
}

func snapshotComputeSpecs(modelName string) map[string]computeSpec {
	computeMu.RLock()
	byField := computes[modelName]
	specs := make(map[string]computeSpec, len(byField))
	for fieldName, spec := range byField {
		specs[fieldName] = spec
	}
	computeMu.RUnlock()
	return specs
}

// ComputeDeps returns dependency field names for a model's computed fields.
func ComputeDeps(model string) []string {
	computeMu.RLock()
	defer computeMu.RUnlock()
	seen := map[string]struct{}{}
	for _, spec := range computes[model] {
		for _, d := range spec.deps {
			seen[d] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

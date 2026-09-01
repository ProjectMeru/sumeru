package orm

import (
	"context"
	"fmt"
)

// MergeStoredComputes runs stored compute handlers in dependency order and merges results for SQL write.
func MergeStoredComputes(ctx context.Context, modelName string, rec map[string]interface{}) error {
	if rec == nil {
		return nil
	}
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return nil
	}
	order := storedComputeOrder(modelName)
	if len(order) == 0 {
		return nil
	}
	computeMu.RLock()
	byField := computes[modelName]
	specs := make(map[string]computeSpec, len(byField))
	for k, v := range byField {
		specs[k] = v
	}
	computeMu.RUnlock()
	for _, field := range order {
		spec, ok := specs[field]
		if !ok || spec.fn == nil {
			continue
		}
		val, err := spec.fn(ctx, rec)
		if err != nil {
			return fmt.Errorf("compute %s.%s: %w", modelName, field, err)
		}
		rec[field] = val
	}
	return nil
}

func storedComputeOrder(modelName string) []string {
	computeMu.RLock()
	byField := computes[modelName]
	if len(byField) == 0 {
		computeMu.RUnlock()
		return nil
	}
	fields := make([]string, 0, len(byField))
	for f := range byField {
		fields = append(fields, f)
	}
	deps := map[string][]string{}
	for f, spec := range byField {
		deps[f] = append([]string(nil), spec.deps...)
	}
	computeMu.RUnlock()
	return topoSort(fields, deps)
}

func topoSort(fields []string, deps map[string][]string) []string {
	fieldSet := map[string]bool{}
	for _, f := range fields {
		fieldSet[f] = true
	}
	inDegree := map[string]int{}
	rev := map[string][]string{}
	for _, f := range fields {
		inDegree[f] = 0
	}
	for _, f := range fields {
		for _, d := range deps[f] {
			if !fieldSet[d] {
				continue
			}
			inDegree[f]++
			rev[d] = append(rev[d], f)
		}
	}
	var queue []string
	for _, f := range fields {
		if inDegree[f] == 0 {
			queue = append(queue, f)
		}
	}
	var out []string
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		out = append(out, n)
		for _, m := range rev[n] {
			inDegree[m]--
			if inDegree[m] == 0 {
				queue = append(queue, m)
			}
		}
	}
	if len(out) != len(fields) {
		return fields
	}
	return out
}

func storedComputeFields(model Model) []string {
	var names []string
	for _, f := range model.Fields() {
		if f.Compute != "" && f.ComputeStore {
			names = append(names, f.Name)
		}
	}
	return names
}

// RejectVirtualWrites returns an error if values touch virtual or readonly-compute fields.
func RejectVirtualWrites(model Model, values map[string]interface{}) error {
	if model == nil || len(values) == 0 {
		return nil
	}
	byName := map[string]FieldDefinition{}
	for _, f := range model.Fields() {
		byName[f.Name] = f
	}
	for k := range values {
		fd, ok := byName[k]
		if !ok {
			continue
		}
		if IsVirtualField(fd) {
			return fmt.Errorf("field %q on %s is read-only", k, model.ModelName())
		}
		if fd.Compute != "" && fd.ComputeStore {
			return fmt.Errorf("field %q on %s is computed", k, model.ModelName())
		}
		if fd.Related != "" {
			return fmt.Errorf("field %q on %s is related", k, model.ModelName())
		}
	}
	return nil
}

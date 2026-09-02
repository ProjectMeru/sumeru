package orm

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ReadGroupWrapper optionally wraps ReadGroupDirect (enterprise pivot cache).
type ReadGroupWrapper func(
	ctx context.Context,
	modelName string,
	domain [][]interface{},
	spec ReadGroupSpec,
	inner func(context.Context, string, [][]interface{}, ReadGroupSpec) ([]ReadGroupRow, error),
) ([]ReadGroupRow, error)

var (
	readGroupWrapperMu sync.RWMutex
	readGroupWrapper   ReadGroupWrapper
)

// RegisterReadGroupWrapper sets the optional ReadGroup wrapper.
func RegisterReadGroupWrapper(w ReadGroupWrapper) {
	readGroupWrapperMu.Lock()
	defer readGroupWrapperMu.Unlock()
	readGroupWrapper = w
}

// ClearReadGroupWrapper removes the wrapper (tests).
func ClearReadGroupWrapper() {
	readGroupWrapperMu.Lock()
	defer readGroupWrapperMu.Unlock()
	readGroupWrapper = nil
}

// ReadGroupSpec defines grouping and aggregation for read_group.
type ReadGroupSpec struct {
	GroupBy []string
	Fields  []ReadGroupField
}

// ReadGroupField is one aggregate column.
type ReadGroupField struct {
	Name    string // output key
	Field   string // source field
	Measure string // sum | count
}

// ReadGroupRow is one grouped result row.
type ReadGroupRow map[string]interface{}

// ReadGroup aggregates records by group fields with sum/count measures.
func ReadGroup(ctx context.Context, modelName string, domain [][]interface{}, spec ReadGroupSpec) ([]ReadGroupRow, error) {
	readGroupWrapperMu.RLock()
	w := readGroupWrapper
	readGroupWrapperMu.RUnlock()
	if w != nil {
		return w(ctx, modelName, domain, spec, ReadGroupDirect)
	}
	return ReadGroupDirect(ctx, modelName, domain, spec)
}

// ReadGroupDirect runs aggregation without optional wrappers.
func ReadGroupDirect(ctx context.Context, modelName string, domain [][]interface{}, spec ReadGroupSpec) ([]ReadGroupRow, error) {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return nil, fmt.Errorf("model required")
	}
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return nil, err
	}
	if len(spec.GroupBy) == 0 {
		return nil, fmt.Errorf("groupby required")
	}

	selectParts := []string{}
	groupCols := []string{}
	for i, g := range spec.GroupBy {
		col, err := QuotedColumnForModel(modelName, g)
		if err != nil {
			return nil, err
		}
		alias := fmt.Sprintf("g%d", i)
		selectParts = append(selectParts, fmt.Sprintf("%s AS %s", col, alias))
		groupCols = append(groupCols, col)
	}
	for _, f := range spec.Fields {
		col, err := QuotedColumnForModel(modelName, f.Field)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSpace(f.Name)
		if name == "" {
			name = f.Field
		}
		switch strings.ToLower(strings.TrimSpace(f.Measure)) {
		case "count", "":
			selectParts = append(selectParts, fmt.Sprintf("COUNT(*) AS %s", quoteIdent(name)))
		case "sum":
			selectParts = append(selectParts, fmt.Sprintf("COALESCE(SUM(%s),0) AS %s", col, quoteIdent(name)))
		case "avg":
			selectParts = append(selectParts, fmt.Sprintf("COALESCE(AVG(%s),0) AS %s", col, quoteIdent(name)))
		case "min":
			selectParts = append(selectParts, fmt.Sprintf("MIN(%s) AS %s", col, quoteIdent(name)))
		case "max":
			selectParts = append(selectParts, fmt.Sprintf("MAX(%s) AS %s", col, quoteIdent(name)))
		default:
			return nil, fmt.Errorf("unsupported measure %q", f.Measure)
		}
	}

	where, args, err := BuildWhereWithRecordRules(ctx, uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}
	tbl, err := QuotedTableForModel(modelName)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s GROUP BY %s ORDER BY 1",
		strings.Join(selectParts, ", "), tbl, where, strings.Join(groupCols, ", "))
	rows, err := QueryDB(ContextWithReadReplica(ctx, true)).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out []ReadGroupRow
	for rows.Next() {
		raw := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := ReadGroupRow{}
		for i, c := range cols {
			row[c] = raw[i]
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

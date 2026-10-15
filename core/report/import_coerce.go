package report

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sumeru/core/orm"
)

// CoerceImportValuesForTest exposes import coercion for external tests.
func CoerceImportValuesForTest(ctx context.Context, modelInst orm.Model, vals map[string]interface{}) (map[string]interface{}, []string) {
	return coerceImportValues(ctx, modelInst, vals)
}

func coerceImportValues(ctx context.Context, modelInst orm.Model, vals map[string]interface{}) (map[string]interface{}, []string) {
	fieldMap := map[string]orm.FieldDefinition{}
	for _, f := range modelInst.Fields() {
		fieldMap[f.Name] = f
	}
	out := map[string]interface{}{}
	var errs []string
	for name, raw := range vals {
		def, ok := fieldMap[name]
		if !ok {
			out[name] = raw
			continue
		}
		coerced, err := coerceFieldValue(ctx, def, raw)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", name, err.Error()))
			continue
		}
		out[name] = coerced
	}
	return out, errs
}

func coerceFieldValue(ctx context.Context, def orm.FieldDefinition, raw interface{}) (interface{}, error) {
	switch def.Type {
	case orm.Boolean:
		switch v := raw.(type) {
		case bool:
			return v, nil
		case string:
			s := strings.ToLower(strings.TrimSpace(v))
			if s == "1" || s == "true" || s == "yes" {
				return true, nil
			}
			if s == "0" || s == "false" || s == "no" || s == "" {
				return false, nil
			}
		}
	case orm.Integer:
		switch v := raw.(type) {
		case int64:
			return v, nil
		case int:
			return int64(v), nil
		case string:
			n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer")
			}
			return n, nil
		}
	case orm.Float, orm.Float64, orm.Numeric:
		switch v := raw.(type) {
		case float64:
			return v, nil
		case int64:
			return float64(v), nil
		case string:
			n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number")
			}
			return n, nil
		}
	case orm.Date:
		if s, ok := raw.(string); ok {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
			if err != nil {
				return nil, fmt.Errorf("invalid date")
			}
			return t, nil
		}
	case orm.DateTime:
		if s, ok := raw.(string); ok {
			s = strings.TrimSpace(s)
			for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
				if t, err := time.Parse(layout, s); err == nil {
					return t, nil
				}
			}
			return nil, fmt.Errorf("invalid datetime")
		}
	case orm.Selection:
		if s, ok := raw.(string); ok {
			s = strings.TrimSpace(s)
			for _, opt := range def.Selection {
				if len(opt) >= 2 && (opt[0] == s || opt[1] == s) {
					return opt[0], nil
				}
			}
			return nil, fmt.Errorf("invalid selection")
		}
	case orm.Many2One:
		return resolveMany2One(ctx, def.Relation, raw)
	}
	return raw, nil
}

func resolveMany2One(ctx context.Context, relation string, raw interface{}) (interface{}, error) {
	relation = strings.TrimSpace(relation)
	if relation == "" {
		return raw, nil
	}
	switch v := raw.(type) {
	case int64:
		if 0 >= v {
			return nil, fmt.Errorf("invalid reference")
		}
		return v, nil
	case int:
		if 0 >= v {
			return nil, fmt.Errorf("invalid reference")
		}
		return int64(v), nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil, nil
		}
		if id, err := strconv.ParseInt(s, 10, 64); err == nil && id > 0 {
			return id, nil
		}
		row, err := orm.SearchOne(ctx, relation, map[string]interface{}{"name": s})
		if err != nil {
			return nil, fmt.Errorf("no %s named %q", relation, s)
		}
		if 0 == len(row) {
			return nil, fmt.Errorf("no %s named %q", relation, s)
		}

		id, _ := orm.CoerceInt64(row["id"])
		return id, nil
	}
	return raw, nil
}

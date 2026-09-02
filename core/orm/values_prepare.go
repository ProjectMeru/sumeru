package orm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// WriteOp selects PrepareValues behavior for create vs write.
type WriteOp string

const (
	WriteOpCreate WriteOp = "create"
	WriteOpWrite  WriteOp = "write"
)

// PrepareOptions controls unknown-key handling.
type PrepareOptions struct {
	// StrictUnknown rejects undeclared field keys. When false, unknown keys are dropped silently
	// (historical Update behavior). Create uses StrictUnknown=true.
	StrictUnknown bool
}

// PrepareValues whitelists model fields, coerces types, and validates required fields on create.
// Relational virtual fields (Many2Many, One2Many) are excluded from SQL column maps.
func PrepareValues(model Model, values map[string]interface{}, op WriteOp, opts PrepareOptions) (map[string]interface{}, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}
	if values == nil {
		values = map[string]interface{}{}
	}
	fieldDefs := fieldDefinitionsByName(model)

	out := make(map[string]interface{}, len(values))
	for k, v := range values {
		if k == "id" {
			continue
		}
		fieldDef, ok := fieldDefs[k]
		if !ok {
			if opts.StrictUnknown {
				return nil, fmt.Errorf("unknown field %q on model %s", k, model.ModelName())
			}
			continue
		}
		switch fieldDef.Type {
		case Many2Many, One2Many:
			// Stored via relation tables; not INSERT/UPDATE columns.
			continue
		}
		if IsVirtualField(fieldDef) {
			if opts.StrictUnknown {
				return nil, fmt.Errorf("field %q on %s is read-only", k, model.ModelName())
			}
			continue
		}
		if fieldDef.Related != "" {
			if opts.StrictUnknown {
				return nil, fmt.Errorf("field %q on %s is related", k, model.ModelName())
			}
			continue
		}
		if fieldDef.Compute != "" && fieldDef.ComputeStore {
			if opts.StrictUnknown {
				return nil, fmt.Errorf("field %q on %s is computed", k, model.ModelName())
			}
			continue
		}
		coercedValue, err := coerceFieldValue(fieldDef, v)
		if err != nil {
			if fve, ok := err.(*FieldValidationError); ok {
				return nil, fve
			}
			return nil, fmt.Errorf("field %s: %w", k, err)
		}
		out[k] = coercedValue
		if err := validateFieldRange(fieldDef, coercedValue); err != nil {
			return nil, err
		}
	}

	if op == WriteOpCreate {
		for name, fieldDef := range fieldDefs {
			if fieldDef.Type == Many2Many || fieldDef.Type == One2Many || IsVirtualField(fieldDef) {
				continue
			}
			if !fieldDef.Required {
				continue
			}
			if _, ok := out[name]; ok {
				continue
			}
			if fieldDef.DefaultVal != nil {
				coercedValue, err := coerceFieldValue(fieldDef, fieldDef.DefaultVal)
				if err != nil {
					return nil, fmt.Errorf("field %s default: %w", name, err)
				}
				out[name] = coercedValue
				continue
			}
			return nil, newFieldValidationError(fieldDef, fmt.Sprintf("required field %q missing on model %s", name, model.ModelName()))
		}
	}
	return out, nil
}

func parseFloatFieldValue(raw interface{}) (float64, error) {
	switch typed := raw.(type) {
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case int, int32, int64:
		intValue, _ := CoerceInt64(raw)
		return float64(intValue), nil
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, nil
		}
		parsed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid float %q", typed)
		}
		return parsed, nil
	default:
		asString := strings.TrimSpace(AsString(raw))
		parsed, err := strconv.ParseFloat(asString, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid float %v", raw)
		}
		return parsed, nil
	}
}

func coerceFieldValue(fieldDef FieldDefinition, v interface{}) (interface{}, error) {
	if v == nil {
		if fieldDef.Required {
			return nil, newFieldValidationError(fieldDef, "")
		}
		return nil, nil
	}
	switch fieldDef.Type {
	case Boolean:
		switch typed := v.(type) {
		case bool:
			return typed, nil
		case string:
			s := strings.ToLower(strings.TrimSpace(typed))
			if s == "true" || s == "1" || s == "yes" {
				return true, nil
			}
			if s == "false" || s == "0" || s == "no" || s == "" {
				return false, nil
			}
			return nil, fmt.Errorf("invalid boolean %q", typed)
		case int, int32, int64, float32, float64:
			intValue, _ := CoerceInt64(v)
			return intValue != 0, nil
		default:
			return AsBool(v), nil
		}
	case Integer, Many2One:
		intValue, ok := CoerceInt64(v)
		if !ok {
			s := strings.TrimSpace(AsString(v))
			if s == "" {
				if fieldDef.Required {
					return nil, newFieldValidationError(fieldDef, "")
				}
				return nil, nil
			}
			return nil, fmt.Errorf("invalid integer %v", v)
		}
		return int(intValue), nil
	case Float, Float64, Numeric:
		parsed, err := parseFloatFieldValue(v)
		if err != nil {
			return nil, err
		}
		if parsed == 0 {
			if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
				return nil, nil
			}
		}
		return parsed, nil
	case Selection:
		s := strings.TrimSpace(AsString(v))
		if s == "" {
			if fieldDef.Required {
				return nil, newFieldValidationError(fieldDef, "")
			}
			return "", nil
		}
		if len(fieldDef.Selection) == 0 {
			return s, nil
		}
		for _, opt := range fieldDef.Selection {
			if len(opt) >= 1 && opt[0] == s {
				return s, nil
			}
		}
		return nil, fmt.Errorf("invalid selection %q", s)
	case Char, Text, Json:
		s := strings.TrimSpace(AsString(v))
		if s == "" && fieldDef.Required {
			return nil, newFieldValidationError(fieldDef, "")
		}
		return AsString(v), nil
	case Date, DateTime:
		s := strings.TrimSpace(AsString(v))
		if s == "" {
			if fieldDef.Required {
				return nil, newFieldValidationError(fieldDef, "")
			}
			return nil, nil
		}
		return s, nil
	default:
		return v, nil
	}
}

func validateFieldRange(fieldDef FieldDefinition, v interface{}) error {
	if fieldDef.Min == nil && fieldDef.Max == nil {
		return nil
	}
	if v == nil {
		return nil
	}
	var numericValue float64
	switch typed := v.(type) {
	case int:
		numericValue = float64(typed)
	case int64:
		numericValue = float64(typed)
	case float32:
		numericValue = float64(typed)
	case float64:
		numericValue = typed
	default:
		return nil
	}
	if fieldDef.Min != nil && numericValue < *fieldDef.Min {
		return newFieldValidationError(fieldDef, fmt.Sprintf("value %v below minimum %v", numericValue, *fieldDef.Min))
	}
	if fieldDef.Max != nil && numericValue > *fieldDef.Max {
		return newFieldValidationError(fieldDef, fmt.Sprintf("value %v above maximum %v", numericValue, *fieldDef.Max))
	}
	return nil
}

func applySpecialDefaults(ctx context.Context, model Model, fieldDefs map[string]FieldDefinition, out map[string]interface{}) error {
	for name, fieldDef := range fieldDefs {
		if _, ok := out[name]; ok {
			continue
		}
		if fieldDef.DefaultVal == nil {
			continue
		}
		defaultToken, ok := fieldDef.DefaultVal.(string)
		if !ok {
			continue
		}
		switch strings.TrimSpace(defaultToken) {
		case "current_user":
			uid := SecurityUID(ctx)
			if uid > 0 {
				out[name] = uid
			}
		case "current_company":
			uid := SecurityUID(ctx)
			if uid > 0 {
				if companyID := ActiveCompanyIDForUser(ctx, uid); companyID > 0 {
					out[name] = int(companyID)
				}
			}
		case "uuid":
			if fieldDef.Type == Char {
				out[name] = newUUID()
			}
		}
	}
	return nil
}

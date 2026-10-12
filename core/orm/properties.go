package orm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	maxPropertyBytes = 64 * 1024
	maxPropertyKeys  = 50
)

var propertyKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// PropertyDef is one schema-lite key stored on sys.property.definition.
type PropertyDef struct {
	Key       string
	ValueType string
	Selection string
	Relation  string
	Required  bool
}

// PropertyDefLoader resolves definitions for a model field. Tests replace it.
var PropertyDefLoader func(model, field string) (map[string]PropertyDef, error)

func loadPropertyDefs(ctx context.Context, modelName, fieldName string) (map[string]PropertyDef, error) {
	if PropertyDefLoader != nil {
		return PropertyDefLoader(modelName, fieldName)
	}
	out := map[string]PropertyDef{}
	if DB == nil {
		return out, nil
	}
	if _, ok := Registry["sys.property.definition"]; !ok {
		return out, nil
	}
	table, err := QuotedTableName("sys.property.definition")
	if err != nil {
		return nil, err
	}
	rows, err := DB.QueryContext(ctx, `SELECT key, value_type, selection, relation, required FROM `+table+` WHERE model = $1 AND field_name = $2`, modelName, fieldName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var valueType, selection, relation sql.NullString
		var required sql.NullBool
		if err := rows.Scan(&key, &valueType, &selection, &relation, &required); err != nil {
			return nil, err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = PropertyDef{
			Key:       key,
			ValueType: strings.TrimSpace(valueType.String),
			Selection: selection.String,
			Relation:  relation.String,
			Required:  required.Bool,
		}
	}
	return out, rows.Err()
}

func coerceProperties(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	var raw []byte
	switch typed := v.(type) {
	case string:
		s := strings.TrimSpace(typed)
		if s == "" || s == "null" {
			return "", nil
		}
		raw = []byte(s)
	case []byte:
		raw = typed
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return "", fmt.Errorf("properties: %w", err)
		}
		raw = encoded
	}
	if len(raw) > maxPropertyBytes {
		return "", fmt.Errorf("properties payload exceeds %d bytes", maxPropertyBytes)
	}
	if string(raw) == "null" {
		return "", nil
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", fmt.Errorf("properties must be a JSON object")
	}
	if obj == nil {
		return "{}", nil
	}
	if len(obj) > maxPropertyKeys {
		return "", fmt.Errorf("properties exceed %d keys", maxPropertyKeys)
	}
	for key := range obj {
		if err := validatePropertyKey(key); err != nil {
			return "", err
		}
	}
	canonical, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	if len(canonical) > maxPropertyBytes {
		return "", fmt.Errorf("properties payload exceeds %d bytes", maxPropertyBytes)
	}
	return string(canonical), nil
}

func validatePropertyKey(key string) error {
	switch key {
	case "__proto__", "constructor", "prototype":
		return fmt.Errorf("properties key %q is not allowed", key)
	}
	if !propertyKeyPattern.MatchString(key) {
		return fmt.Errorf("properties key %q is not allowed", key)
	}
	return nil
}

func parsePropertiesObject(raw interface{}) (map[string]interface{}, error) {
	canonical, err := coerceProperties(raw)
	if err != nil {
		return nil, err
	}
	if canonical == "" {
		return map[string]interface{}{}, nil
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(canonical), &obj); err != nil {
		return nil, err
	}
	if obj == nil {
		obj = map[string]interface{}{}
	}
	return obj, nil
}

// MergePropertiesWrite checks defined keys and keeps values whose definition was removed.
func MergePropertiesWrite(ctx context.Context, modelName string, before, prepared map[string]interface{}) error {
	inst := RegistryModel(modelName)
	if inst == nil {
		return nil
	}
	for _, field := range inst.Fields() {
		if field.Type != Properties {
			continue
		}
		incomingRaw, ok := prepared[field.Name]
		if !ok {
			continue
		}
		incoming, err := parsePropertiesObject(incomingRaw)
		if err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
		previous := map[string]interface{}{}
		if before != nil {
			previous, err = parsePropertiesObject(before[field.Name])
			if err != nil {
				previous = map[string]interface{}{}
			}
		}
		defs, err := loadPropertyDefs(ctx, modelName, field.Name)
		if err != nil {
			return err
		}
		merged := map[string]interface{}{}
		for key, value := range previous {
			if _, defined := defs[key]; !defined {
				merged[key] = value
			}
		}
		for key, value := range incoming {
			def, defined := defs[key]
			if !defined {
				return fmt.Errorf("field %s: unknown property %q", field.Name, key)
			}
			checked, err := checkPropertyValue(def, value)
			if err != nil {
				return fmt.Errorf("field %s: %w", field.Name, err)
			}
			merged[key] = checked
		}
		for _, def := range defs {
			if !def.Required {
				continue
			}
			if _, ok := merged[def.Key]; !ok {
				return fmt.Errorf("field %s: required property %q missing", field.Name, def.Key)
			}
		}
		if len(merged) > maxPropertyKeys {
			return fmt.Errorf("field %s: properties exceed %d keys", field.Name, maxPropertyKeys)
		}
		encoded, err := json.Marshal(merged)
		if err != nil {
			return err
		}
		prepared[field.Name] = string(encoded)
	}
	return nil
}

func checkPropertyValue(def PropertyDef, value interface{}) (interface{}, error) {
	switch def.ValueType {
	case "char", "":
		return strings.TrimSpace(AsString(value)), nil
	case "integer", "many2one":
		n, ok := CoerceInt64(value)
		if !ok {
			return nil, fmt.Errorf("property %q must be an integer", def.Key)
		}
		if def.ValueType == "many2one" && n <= 0 {
			return nil, fmt.Errorf("property %q must be a positive integer", def.Key)
		}
		return n, nil
	case "boolean":
		return AsBool(value), nil
	case "date":
		s := strings.TrimSpace(AsString(value))
		if len(s) < 10 {
			return nil, fmt.Errorf("property %q must be a date", def.Key)
		}
		return s[:10], nil
	case "selection":
		s := strings.TrimSpace(AsString(value))
		allowed := map[string]struct{}{}
		for _, part := range strings.Split(def.Selection, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			key := part
			if i := strings.IndexByte(part, ':'); i >= 0 {
				key = part[:i]
			}
			allowed[key] = struct{}{}
		}
		if _, ok := allowed[s]; !ok {
			return nil, fmt.Errorf("property %q value %q is not in the selection", def.Key, s)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("property %q has unknown type %q", def.Key, def.ValueType)
	}
}

func propertiesDomainClause(modelName, fieldExpr, op string, value interface{}, placeholder int) (string, []interface{}, int, error) {
	field, key, ok := strings.Cut(fieldExpr, ".")
	if !ok || field == "" || key == "" {
		return "", nil, placeholder, fmt.Errorf("invalid property domain %q", fieldExpr)
	}
	if err := validatePropertyKey(key); err != nil {
		return "", nil, placeholder, err
	}
	fd := FieldDef(modelName, field)
	if fd == nil || fd.Type != Properties {
		return "", nil, placeholder, fmt.Errorf("field %q on %s is not properties", field, modelName)
	}
	defs, err := loadPropertyDefs(context.Background(), modelName, field)
	if err != nil {
		return "", nil, placeholder, err
	}
	def, defined := defs[key]
	if !defined {
		return "", nil, placeholder, fmt.Errorf("undefined property key %q", key)
	}
	col, err := QuotedColumnForModel(modelName, field)
	if err != nil {
		return "", nil, placeholder, err
	}
	expr := fmt.Sprintf("(%s->>'%s')", col, key)
	op = strings.TrimSpace(strings.ToLower(op))
	switch def.ValueType {
	case "integer", "many2one":
		switch op {
		case "=", "!=", ">", ">=", "<", "<=":
			cmp := strings.ToUpper(op)
			if op == "!=" {
				return fmt.Sprintf("(%s IS DISTINCT FROM $%d::text)", expr, placeholder), []interface{}{fmt.Sprint(value)}, placeholder + 1, nil
			}
			if op == "=" {
				return fmt.Sprintf("%s = $%d", expr, placeholder), []interface{}{fmt.Sprint(value)}, placeholder + 1, nil
			}
			safe := fmt.Sprintf("(CASE WHEN %s ~ '^-?[0-9]+$' THEN %s::bigint END) %s $%d", expr, expr, cmp, placeholder)
			n, ok := CoerceInt64(value)
			if !ok {
				return "", nil, placeholder, fmt.Errorf("property %q comparison needs an integer", key)
			}
			return safe, []interface{}{n}, placeholder + 1, nil
		default:
			return "", nil, placeholder, fmt.Errorf("operator %q is not valid for property %q", op, key)
		}
	default:
		switch op {
		case "=":
			return fmt.Sprintf("%s = $%d", expr, placeholder), []interface{}{value}, placeholder + 1, nil
		case "!=":
			return fmt.Sprintf("(%s IS DISTINCT FROM $%d)", expr, placeholder), []interface{}{value}, placeholder + 1, nil
		case "ilike", "like":
			if def.ValueType != "char" && def.ValueType != "" {
				return "", nil, placeholder, fmt.Errorf("operator %q is not valid for property %q", op, key)
			}
			return fmt.Sprintf("%s ILIKE $%d", expr, placeholder), []interface{}{value}, placeholder + 1, nil
		default:
			return "", nil, placeholder, fmt.Errorf("operator %q is not valid for property %q", op, key)
		}
	}
}

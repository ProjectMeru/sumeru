package orm

import (
	"context"
	"fmt"
	"strings"
)

// tableExists reports whether a physical table exists in the public schema.
func tableExists(ctx context.Context, tableName string) (bool, error) {
	if DB == nil {
		return false, nil
	}
	var count int
	err := DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1
	`, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// quoteSQLString escapes a string for use inside single-quoted SQL literals.
func quoteSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// sqlDefaultLiteral formats a Go default value as a SQL literal for DDL DEFAULT clauses.
func sqlDefaultLiteral(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	switch t := v.(type) {
	case bool:
		if t {
			return "TRUE", true
		}
		return "FALSE", true
	case int:
		return fmt.Sprintf("%d", t), true
	case int64:
		return fmt.Sprintf("%d", t), true
	case float32:
		return fmt.Sprintf("%g", float64(t)), true
	case float64:
		return fmt.Sprintf("%g", t), true
	case string:
		if strings.Contains(t, ";") || strings.Contains(t, "--") || strings.Contains(t, "/*") {
			return "", false
		}
		return "'" + quoteSQLString(t) + "'", true
	default:
		return "'" + quoteSQLString(fmt.Sprint(t)) + "'", true
	}
}

// isRuntimeDefaultToken reports defaults applied in Go at insert time, not as SQL DEFAULT.
func isRuntimeDefaultToken(v interface{}) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	switch strings.TrimSpace(s) {
	case "uuid", "current_user", "current_company":
		return true
	default:
		return false
	}
}

// columnDefaultClause returns a SQL DEFAULT suffix for DDL, or empty when no SQL default applies.
func columnDefaultClause(modelName string, field FieldDefinition) (string, error) {
	defaultVal := field.DefaultVal
	if isRuntimeDefaultToken(defaultVal) {
		return "", nil
	}
	if defaultVal == nil {
		return "", nil
	}
	if defaultString, ok := defaultVal.(string); ok {
		if strings.Contains(defaultString, ";") || strings.Contains(defaultString, "--") || strings.Contains(defaultString, "/*") {
			return "", fmt.Errorf("unsafe string default on %s.%s", modelName, field.Name)
		}
	}
	switch defaultVal.(type) {
	case string, bool, int, int64, float64:
		if literal, ok := sqlDefaultLiteral(defaultVal); ok {
			return " DEFAULT " + literal, nil
		}
	}
	return "", nil
}

// formatAddColumnDefinition builds the SQL column definition fragment for ALTER TABLE ... ADD COLUMN.
func formatAddColumnDefinition(field FieldDefinition, baseType string) string {
	defaultVal := field.DefaultVal
	if isRuntimeDefaultToken(defaultVal) {
		defaultVal = nil
	}
	if field.Type == Boolean {
		if defaultVal == true {
			return baseType + " NOT NULL DEFAULT TRUE"
		}
		if defaultVal == false {
			return baseType + " NOT NULL DEFAULT FALSE"
		}
		if literal, ok := sqlDefaultLiteral(defaultVal); ok {
			return baseType + " DEFAULT " + literal
		}
		return baseType
	}
	if literal, ok := sqlDefaultLiteral(defaultVal); ok {
		if field.Required {
			return baseType + " NOT NULL DEFAULT " + literal
		}
		return baseType + " DEFAULT " + literal
	}
	return baseType
}

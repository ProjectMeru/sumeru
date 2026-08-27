package orm

import (
	"fmt"
	"strings"
)

// tableExists reports whether a physical table exists in the public schema.
func tableExists(tableName string) (bool, error) {
	if DB == nil {
		return false, nil
	}
	var count int
	err := DB.QueryRow(`
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

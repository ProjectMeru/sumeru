package orm

import "database/sql"

func scanRowToMap(cols []string, rows *sql.Rows) (map[string]interface{}, error) {
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}
	m := make(map[string]interface{}, len(cols))
	for i, c := range cols {
		m[c] = normalizeScannedValue(vals[i])
	}
	return m, nil
}

// normalizeScannedValue converts driver []byte values (NUMERIC, JSONB) into
// strings so they serialize as plain JSON text instead of base64.
func normalizeScannedValue(v interface{}) interface{} {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

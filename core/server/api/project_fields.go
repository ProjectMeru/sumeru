package api

// projectFields returns rows limited to the requested field names.
// An empty fields list returns rows unchanged.
func projectFields(rows []map[string]interface{}, fields []string) []map[string]interface{} {
	if 0 == len(fields) {
		return rows
	}
	set := map[string]struct{}{}
	for _, f := range fields {
		set[f] = struct{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		m := make(map[string]interface{}, len(fields))
		for k, v := range r {
			if _, ok := set[k]; ok {
				m[k] = v
			}
		}
		out = append(out, m)
	}
	return out
}

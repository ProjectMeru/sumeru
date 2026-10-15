package orm

import "strings"

// BuildListSearchDomain OR-combines ilike clauses for searchable string fields.
func BuildListSearchDomain(modelName string, fieldNames []string, query string) [][]interface{} {
	query = strings.TrimSpace(query)
	if query == "" || 0 == len(fieldNames) {
		return nil
	}
	pattern := "%" + query + "%"
	var leaves [][]interface{}
	for _, f := range fieldNames {
		f = strings.TrimSpace(f)
		if f == "" || f == "id" {
			continue
		}
		if !fieldSearchable(modelName, f) {
			continue
		}
		leaves = append(leaves, []interface{}{f, "ilike", pattern})
	}
	if 0 == len(leaves) {
		return nil
	}
	if len(leaves) == 1 {
		return leaves
	}
	out := make([][]interface{}, 0, len(leaves)*2-1)
	for i := 0; i < len(leaves)-1; i++ {
		out = append(out, []interface{}{"|"})
	}
	out = append(out, leaves...)
	return out
}

func fieldSearchable(modelName, field string) bool {
	m, ok := Registry[modelName]
	if !ok {
		return false
	}
	for _, fd := range m.Fields() {
		if fd.Name != field {
			continue
		}
		switch fd.Type {
		case Char, Text:
			return true
		default:
			return false
		}
	}
	return false
}

// MergeDomains AND-combines two domains (base action domain + optional search overlay).
func MergeDomains(base, extra [][]interface{}) [][]interface{} {
	if 0 == len(extra) {
		return base
	}
	if 0 == len(base) {
		return extra
	}
	out := make([][]interface{}, 0, len(base)+len(extra))
	out = append(out, base...)
	out = append(out, extra...)
	return out
}

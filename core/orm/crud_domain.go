package orm

import (
	"fmt"
	"strings"
)

func buildSearchWhereClause(modelName string, domain [][]interface{}) (string, []interface{}, error) {
	if len(domain) == 0 {
		return "1=1", nil, nil
	}
	// Prefix Polish OR: ["|", ...] repeated (N-1) times then N leaf triples → OR of leaves.
	orLeaves := 0
	for _, d := range domain {
		if len(d) == 1 && fmt.Sprint(d[0]) == "|" {
			orLeaves++
			continue
		}
		break
	}
	if orLeaves > 0 {
		leaves := domain[orLeaves:]
		if len(leaves) != orLeaves+1 {
			return "", nil, fmt.Errorf("invalid OR domain shape")
		}
		var parts []string
		var args []interface{}
		n := 1
		for _, leaf := range leaves {
			frag, a, err := buildSearchWhereClause(modelName, [][]interface{}{leaf})
			if err != nil {
				return "", nil, err
			}
			shifted, err := shiftPlaceholders(frag, n)
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, "("+shifted+")")
			args = append(args, a...)
			n += len(a)
		}
		return strings.Join(parts, " OR "), args, nil
	}
	var parts []string
	var args []interface{}
	n := 1
	for _, d := range domain {
		if len(d) != 3 {
			return "", nil, fmt.Errorf("invalid domain clause %v", d)
		}
		field, ok := d[0].(string)
		if !ok || strings.TrimSpace(field) == "" {
			return "", nil, fmt.Errorf("domain field name")
		}
		op := strings.TrimSpace(strings.ToLower(fmt.Sprint(d[1])))
		col, err := QuotedColumnForModel(modelName, field)
		if err != nil {
			return "", nil, err
		}
		if dateLike, isBool := dateLikeNullCheck(modelName, field, d[2]); isBool {
			switch op {
			case "!=":
				if dateLike {
					parts = append(parts, fmt.Sprintf("%s IS NOT NULL", col))
				} else {
					parts = append(parts, fmt.Sprintf("%s IS NULL", col))
				}
				continue
			case "=":
				if dateLike {
					parts = append(parts, fmt.Sprintf("%s IS NULL", col))
				} else {
					parts = append(parts, fmt.Sprintf("%s IS NOT NULL", col))
				}
				continue
			}
		}
		switch op {
		case "=":
			parts = append(parts, fmt.Sprintf("%s = $%d", col, n))
			args = append(args, d[2])
			n++
		case "!=":
			parts = append(parts, fmt.Sprintf("(%s IS DISTINCT FROM $%d)", col, n))
			args = append(args, d[2])
			n++
		case "in":
			list, ok := d[2].([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("operator in requires array value")
			}
			if len(list) == 0 {
				parts = append(parts, "FALSE")
				continue
			}
			ph := make([]string, len(list))
			for i := range list {
				ph[i] = fmt.Sprintf("$%d", n)
				args = append(args, list[i])
				n++
			}
			parts = append(parts, fmt.Sprintf("%s IN (%s)", col, strings.Join(ph, ",")))
		case "ilike", "like":
			parts = append(parts, fmt.Sprintf("%s ILIKE $%d", col, n))
			args = append(args, d[2])
			n++
		case "=like":
			parts = append(parts, fmt.Sprintf("%s LIKE $%d", col, n))
			args = append(args, d[2])
			n++
		case ">", ">=", "<", "<=":
			parts = append(parts, fmt.Sprintf("%s %s $%d", col, strings.ToUpper(op), n))
			args = append(args, d[2])
			n++
		case "not in":
			list, ok := d[2].([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("operator not in requires array value")
			}
			if len(list) == 0 {
				parts = append(parts, "TRUE")
				continue
			}
			ph := make([]string, len(list))
			for i := range list {
				ph[i] = fmt.Sprintf("$%d", n)
				args = append(args, list[i])
				n++
			}
			parts = append(parts, fmt.Sprintf("%s NOT IN (%s)", col, strings.Join(ph, ",")))
		default:
			return "", nil, fmt.Errorf("unsupported domain operator %q", op)
		}
	}
	return strings.Join(parts, " AND "), args, nil
}

// dateLikeNullCheck detects date unset/set domains: ("field", "!=", false) means IS NOT NULL.
func dateLikeNullCheck(modelName, fieldName string, value interface{}) (isSetCheck bool, ok bool) {
	b, ok := value.(bool)
	if !ok {
		return false, false
	}
	inst, has := Registry[modelName]
	if !has || inst == nil {
		return false, false
	}
	for _, f := range inst.Fields() {
		if f.Name != fieldName {
			continue
		}
		if f.Type != Date && f.Type != DateTime {
			return false, false
		}
		return !b, true
	}
	return false, false
}

// buildAndWhereClauses ANDs independent domain parts, each compiled separately
// so OR prefixes inside a part remain valid (global AND (g1 OR g2)).
func buildAndWhereClauses(modelName string, parts [][][]interface{}) (string, []interface{}, error) {
	if len(parts) == 0 {
		return "1=1", nil, nil
	}
	var sqlParts []string
	var args []interface{}
	n := 1
	for _, p := range parts {
		frag, a, err := buildSearchWhereClause(modelName, p)
		if err != nil {
			return "", nil, err
		}
		shifted, err := shiftPlaceholders(frag, n)
		if err != nil {
			return "", nil, err
		}
		sqlParts = append(sqlParts, "("+shifted+")")
		args = append(args, a...)
		n += len(a)
	}
	return strings.Join(sqlParts, " AND "), args, nil
}

func shiftPlaceholders(frag string, start int) (string, error) {
	// Replace from highest index downward to avoid $1 colliding with $10.
	max := 0
	for i := 1; i <= 256; i++ {
		if strings.Contains(frag, fmt.Sprintf("$%d", i)) {
			max = i
		}
	}
	out := frag
	for i := max; i >= 1; i-- {
		out = strings.ReplaceAll(out, fmt.Sprintf("$%d", i), fmt.Sprintf("$$TMP%d$$", i))
	}
	for i := max; i >= 1; i-- {
		out = strings.ReplaceAll(out, fmt.Sprintf("$$TMP%d$$", i), fmt.Sprintf("$%d", start+i-1))
	}
	return out, nil
}

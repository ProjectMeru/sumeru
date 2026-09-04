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
	orLeaves, leaves, ok := splitDomainORPrefix(domain)
	if orLeaves > 0 {
		if !ok {
			return "", nil, fmt.Errorf("invalid OR domain shape")
		}
		orDomains := make([][][]interface{}, len(leaves))
		for i, leaf := range leaves {
			orDomains[i] = [][]interface{}{leaf}
		}
		return joinShiftedWhereFragments(" OR ", modelName, orDomains, 1)
	}
	var parts []string
	var args []interface{}
	placeholderIndex := 1
	for _, clause := range domain {
		if len(clause) != 3 {
			return "", nil, fmt.Errorf("invalid domain clause %v", clause)
		}
		field, ok := clause[0].(string)
		if !ok || strings.TrimSpace(field) == "" {
			return "", nil, fmt.Errorf("domain field name")
		}
		op := strings.TrimSpace(strings.ToLower(fmt.Sprint(clause[1])))
		if frag, fragArgs, handled, err := buildFalsyDomainClause(modelName, field, op, clause[2]); handled {
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, frag)
			args = append(args, fragArgs...)
			continue
		}
		col, err := QuotedColumnForModel(modelName, field)
		if err != nil {
			return "", nil, err
		}
		switch op {
		case "=":
			parts = append(parts, fmt.Sprintf("%s = $%d", col, placeholderIndex))
			args = append(args, clause[2])
			placeholderIndex++
		case "!=":
			parts = append(parts, fmt.Sprintf("(%s IS DISTINCT FROM $%d)", col, placeholderIndex))
			args = append(args, clause[2])
			placeholderIndex++
		case "in":
			list, ok := clause[2].([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("operator in requires array value")
			}
			listClause, listArgs, nextIndex, err := appendDomainListClause(col, "in", list, placeholderIndex)
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, listClause)
			args = append(args, listArgs...)
			placeholderIndex = nextIndex
		case "ilike", "like":
			parts = append(parts, fmt.Sprintf("%s ILIKE $%d", col, placeholderIndex))
			args = append(args, clause[2])
			placeholderIndex++
		case "=like":
			parts = append(parts, fmt.Sprintf("%s LIKE $%d", col, placeholderIndex))
			args = append(args, clause[2])
			placeholderIndex++
		case ">", ">=", "<", "<=":
			parts = append(parts, fmt.Sprintf("%s %s $%d", col, strings.ToUpper(op), placeholderIndex))
			args = append(args, clause[2])
			placeholderIndex++
		case "not in":
			list, ok := clause[2].([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("operator not in requires array value")
			}
			listClause, listArgs, nextIndex, err := appendDomainListClause(col, "not in", list, placeholderIndex)
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, listClause)
			args = append(args, listArgs...)
			placeholderIndex = nextIndex
		default:
			return "", nil, fmt.Errorf("unsupported domain operator %q", op)
		}
	}
	return strings.Join(parts, " AND "), args, nil
}

func appendDomainListClause(col, op string, list []interface{}, placeholderIndex int) (clause string, args []interface{}, nextIndex int, err error) {
	if op == "in" {
		if len(list) == 0 {
			return "FALSE", nil, placeholderIndex, nil
		}
		placeholders := make([]string, len(list))
		for i, item := range list {
			placeholders[i] = fmt.Sprintf("$%d", placeholderIndex)
			args = append(args, item)
			placeholderIndex++
		}
		return fmt.Sprintf("%s IN (%s)", col, strings.Join(placeholders, ",")), args, placeholderIndex, nil
	}
	if len(list) == 0 {
		return "TRUE", nil, placeholderIndex, nil
	}
	placeholders := make([]string, len(list))
	for i, item := range list {
		placeholders[i] = fmt.Sprintf("$%d", placeholderIndex)
		args = append(args, item)
		placeholderIndex++
	}
	return fmt.Sprintf("%s NOT IN (%s)", col, strings.Join(placeholders, ",")), args, placeholderIndex, nil
}

func joinShiftedWhereFragments(separator, modelName string, domains [][][]interface{}, startPlaceholder int) (string, []interface{}, error) {
	if len(domains) == 0 {
		return "1=1", nil, nil
	}
	var parts []string
	var args []interface{}
	placeholderIndex := startPlaceholder
	for _, domain := range domains {
		whereFragment, fragmentArgs, err := buildSearchWhereClause(modelName, domain)
		if err != nil {
			return "", nil, err
		}
		shifted, err := shiftPlaceholders(whereFragment, placeholderIndex)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, "("+shifted+")")
		args = append(args, fragmentArgs...)
		placeholderIndex += len(fragmentArgs)
	}
	return strings.Join(parts, separator), args, nil
}

// buildAndWhereClauses ANDs independent domain parts, each compiled separately
// so OR prefixes inside a part remain valid (global AND (g1 OR g2)).
func buildAndWhereClauses(modelName string, parts [][][]interface{}) (string, []interface{}, error) {
	return joinShiftedWhereFragments(" AND ", modelName, parts, 1)
}

func shiftPlaceholders(whereFragment string, start int) (string, error) {
	// Replace from highest index downward to avoid $1 colliding with $10.
	max := 0
	for i := 1; i <= 256; i++ {
		if strings.Contains(whereFragment, fmt.Sprintf("$%d", i)) {
			max = i
		}
	}
	out := whereFragment
	for i := max; i >= 1; i-- {
		out = strings.ReplaceAll(out, fmt.Sprintf("$%d", i), fmt.Sprintf("$$TMP%d$$", i))
	}
	for i := max; i >= 1; i-- {
		out = strings.ReplaceAll(out, fmt.Sprintf("$$TMP%d$$", i), fmt.Sprintf("$%d", start+i-1))
	}
	return out, nil
}

package eval

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SafeEval parses literals, ref tuples, domain lists, and datetime helpers for XML eval=.
func SafeEval(raw string) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if v, ok, err := parseDomainList(raw); ok {
		return v, err
	}
	if strings.HasPrefix(raw, "ref(") && strings.HasSuffix(raw, ")") {
		refName := strings.TrimSpace(raw[4 : len(raw)-1])
		refName = strings.Trim(refName, `"'`)
		if refName == "" {
			return nil, fmt.Errorf("empty ref() in eval")
		}
		return []interface{}{"ref", refName}, nil
	}
	if strings.HasPrefix(raw, "time.") {
		return evalTimeHelper(raw)
	}
	return safeEvalLiteral(raw)
}

func safeEvalLiteral(raw string) (interface{}, error) {
	lower := strings.ToLower(raw)
	switch lower {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	case "none", "null":
		return nil, nil
	}
	if strings.HasPrefix(raw, `"`) && strings.HasSuffix(raw, `"`) && len(raw) >= 2 {
		return raw[1 : len(raw)-1], nil
	}
	if strings.HasPrefix(raw, `'`) && strings.HasSuffix(raw, `'`) && len(raw) >= 2 {
		return raw[1 : len(raw)-1], nil
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return n, nil
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f, nil
	}
	if strings.HasPrefix(raw, "(") && strings.HasSuffix(raw, ")") {
		inner := strings.TrimSpace(raw[1 : len(raw)-1])
		parts := splitRefTuple(inner)
		if len(parts) > 0 {
			return parts, nil
		}
	}
	if strings.Contains(raw, ",") {
		parts := splitRefTuple(raw)
		if len(parts) > 0 {
			return parts, nil
		}
	}
	return raw, nil
}

func parseDomainList(raw string) ([][]interface{}, bool, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "[") || !strings.HasSuffix(raw, "]") {
		return nil, false, nil
	}
	inner := strings.TrimSpace(raw[1 : len(raw)-1])
	if inner == "" {
		return [][]interface{}{}, true, nil
	}
	var domain [][]interface{}
	depth := 0
	start := -1
	for i, ch := range inner {
		switch ch {
		case '(':
			if depth == 0 {
				start = i
			}
			depth++
		case ')':
			depth--
			if depth == 0 && start >= 0 {
				triple, err := parseDomainTriple(inner[start : i+1])
				if err != nil {
					return nil, true, err
				}
				domain = append(domain, triple)
				start = -1
			}
		}
	}
	if len(domain) == 0 {
		return nil, false, nil
	}
	return domain, true, nil
}

func parseDomainTriple(raw string) ([]interface{}, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "(") || !strings.HasSuffix(raw, ")") {
		return nil, fmt.Errorf("invalid domain triple %q", raw)
	}
	parts := splitEvalList(raw[1 : len(raw)-1])
	if len(parts) < 3 {
		return nil, fmt.Errorf("domain triple needs 3 parts: %q", raw)
	}
	field, err := safeEvalLiteral(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, err
	}
	op, err := safeEvalLiteral(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}
	value, err := evalDomainValue(strings.TrimSpace(parts[2]))
	if err != nil {
		return nil, err
	}
	return []interface{}{field, op, value}, nil
}

func evalDomainValue(raw string) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		inner := strings.TrimSpace(raw[1 : len(raw)-1])
		if inner == "" {
			return []interface{}{}, nil
		}
		var list []interface{}
		for _, part := range splitEvalList(inner) {
			v, err := safeEvalLiteral(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
		return list, nil
	}
	return safeEvalLiteral(raw)
}

func splitEvalList(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func evalTimeHelper(raw string) (interface{}, error) {
	switch strings.TrimSpace(raw) {
	case "time.today()":
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	case "time.now()":
		return time.Now().UTC(), nil
	default:
		return nil, fmt.Errorf("unsupported time helper %q", raw)
	}
}

func splitRefTuple(s string) []interface{} {
	var out []interface{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if v, err := SafeEval(part); err == nil {
			out = append(out, v)
		} else {
			out = append(out, part)
		}
	}
	return out
}

// MustSafeEval is like SafeEval but returns an error for unsupported complex expressions.
func MustSafeEval(raw string) (interface{}, error) {
	v, err := SafeEval(raw)
	if err != nil {
		return nil, err
	}
	if strings.ContainsAny(raw, "+-*/") && !strings.HasPrefix(strings.TrimSpace(raw), "(") && !strings.HasPrefix(strings.TrimSpace(raw), "[") {
		return nil, fmt.Errorf("unsupported eval expression %q", raw)
	}
	return v, nil
}

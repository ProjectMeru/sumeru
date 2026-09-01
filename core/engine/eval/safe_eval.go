package eval

import (
	"fmt"
	"strconv"
	"strings"
)

// SafeEval parses a restricted literal/expression subset for XML data eval= attributes.
func SafeEval(raw string) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
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
	if strings.ContainsAny(raw, "+-*/") && !strings.HasPrefix(strings.TrimSpace(raw), "(") {
		return nil, fmt.Errorf("unsupported eval expression %q", raw)
	}
	return v, nil
}

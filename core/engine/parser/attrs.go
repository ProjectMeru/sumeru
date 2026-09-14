package parser

import (
	"fmt"
	"strings"
)

// ParseXMLBoolAttr parses a boolean XML attribute (empty → false; only true/false allowed).
func ParseXMLBoolAttr(name, raw string) (bool, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return false, nil
	}
	if strings.EqualFold(v, "true") {
		return true, nil
	}
	if strings.EqualFold(v, "false") {
		return false, nil
	}
	return false, fmt.Errorf("invalid %s=%q (only \"true\" or \"false\" allowed)", name, v)
}

func isLegacyBoolAlias(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "0", "1", "yes", "no", "on", "off":
		return true
	default:
		return false
	}
}

// ParseModifierAttr splits invisible/readonly/required into a literal or expression.
func ParseModifierAttr(name, raw string) (literal bool, truthy bool, expr string, err error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return true, false, "", nil
	}
	if strings.EqualFold(s, "true") {
		return true, true, "", nil
	}
	if strings.EqualFold(s, "false") {
		return true, false, "", nil
	}
	if isLegacyBoolAlias(s) {
		return false, false, "", fmt.Errorf("invalid %s=%q (only \"true\" or \"false\" allowed)", name, s)
	}
	return false, false, s, nil
}

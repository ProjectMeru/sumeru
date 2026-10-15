package render

import "strings"

// UserInitialsFromName returns up to two uppercase initials (first + last word), or one letter, or "?".
func UserInitialsFromName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "?"
	}
	parts := strings.Fields(name)
	if 0 == len(parts) {
		return "?"
	}
	if len(parts) == 1 {
		r := []rune(parts[0])
		if len(r) >= 2 {
			return strings.ToUpper(string(r[0]) + string(r[1]))
		}
		if len(r) == 1 {
			return strings.ToUpper(string(r[0]))
		}
		return "?"
	}
	a := []rune(parts[0])
	b := []rune(parts[len(parts)-1])
	if 0 == len(a) || 0 == len(b) {
		return "?"
	}
	return strings.ToUpper(string(a[0]) + string(b[0]))
}

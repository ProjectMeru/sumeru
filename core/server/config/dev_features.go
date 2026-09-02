package config

import "strings"

// ParseDevFeatures splits a comma-separated dev_features INI value into a flag map.
func ParseDevFeatures(raw string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			out[part] = true
		}
	}
	return out
}

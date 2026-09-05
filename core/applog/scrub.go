package applog

import "strings"

// RedactedPlaceholder replaces secret values in logs and dumps.
const RedactedPlaceholder = "***"

// ScrubMap returns a copy of fields with secret field values redacted.
func ScrubMap(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}
	scrubbed := make(map[string]interface{}, len(fields))
	for fieldName, value := range fields {
		scrubbed[fieldName] = ScrubValue(fieldName, value)
	}
	return scrubbed
}

// ScrubValue redacts value when fieldName is secret; walks nested maps and slices.
func ScrubValue(fieldName string, value any) any {
	if IsSecretKey(fieldName) {
		return RedactedPlaceholder
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		return ScrubMap(typed)
	case map[string]string:
		scrubbed := make(map[string]interface{}, len(typed))
		for nestedName, nestedValue := range typed {
			scrubbed[nestedName] = ScrubValue(nestedName, nestedValue)
		}
		return scrubbed
	case []map[string]interface{}:
		scrubbed := make([]interface{}, len(typed))
		for i := range typed {
			scrubbed[i] = ScrubMap(typed[i])
		}
		return scrubbed
	case []interface{}:
		scrubbed := make([]interface{}, len(typed))
		for i := range typed {
			scrubbed[i] = ScrubValue(fieldName, typed[i])
		}
		return scrubbed
	default:
		return value
	}
}

// IsSecretKey reports whether a field name must never be logged in cleartext.
// "sid" matches the full field name only (avoids false positives like "inside").
func IsSecretKey(fieldName string) bool {
	normalized := strings.ToLower(strings.TrimSpace(fieldName))
	return normalized == "sid" || (normalized != "" && containsSecretKeyword(normalized))
}

// TextContainsSecretKeyword reports whether free text (e.g. SQL) mentions a secret term.
func TextContainsSecretKeyword(text string) bool {
	lowered := strings.ToLower(text)
	return strings.Contains(lowered, "sid") || containsSecretKeyword(lowered)
}

func containsSecretKeyword(haystack string) bool {
	for _, keyword := range secretKeywords {
		if strings.Contains(haystack, keyword) {
			return true
		}
	}
	return false
}

// secretKeywords match as substrings of field names and of free text (including totp*).
var secretKeywords = []string{
	"password", "token", "secret", "authorization", "cookie",
	"session", "api_key", "apikey", "key_hash", "csrf", "totp",
}

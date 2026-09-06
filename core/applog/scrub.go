package applog

import "strings"

const RedactedPlaceholder = "***"

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

// IsSecretKey: "sid" is exact-match only (avoids "inside").
func IsSecretKey(fieldName string) bool {
	normalized := strings.ToLower(strings.TrimSpace(fieldName))
	return normalized == "sid" || (normalized != "" && containsSecretKeyword(normalized))
}

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

var secretKeywords = []string{
	"password", "token", "secret", "authorization", "cookie",
	"session", "api_key", "apikey", "key_hash", "csrf", "totp",
}

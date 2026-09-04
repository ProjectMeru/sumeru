package orm

import (
	"context"
	"strings"
)

// GetConfigParam returns the value for key, or defaultVal when missing or empty.
func GetConfigParam(ctx context.Context, key, defaultVal string) string {
	return GetConfig(ctx, key, defaultVal)
}

// SetConfigParam upserts a sys.config.parameter row by key.
func SetConfigParam(ctx context.Context, key, value string) error {
	return SetConfig(ctx, key, value)
}

// ConfigParamBool parses a config parameter as boolean (true/1/t/yes).
func ConfigParamBool(ctx context.Context, key string, defaultVal bool) bool {
	raw := strings.ToLower(GetConfigParam(ctx, key, ""))
	if raw == "" {
		return defaultVal
	}
	return raw == "true" || raw == "1" || raw == "t" || raw == "yes"
}

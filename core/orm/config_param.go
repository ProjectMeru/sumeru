package orm

import (
	"context"
	"strings"
)

const configParamModel = "sys.config.parameter"

// GetConfigParam returns the value for key, or defaultVal when missing or empty.
func GetConfigParam(ctx context.Context, key, defaultVal string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return defaultVal
	}
	row, err := SearchOne(ctx, configParamModel, map[string]interface{}{"key": key})
	if err != nil {
		return defaultVal
	}
	val := strings.TrimSpace(AsString(row["value"]))
	if val == "" {
		return defaultVal
	}
	return val
}

// SetConfigParam upserts a sys.config.parameter row by key.
func SetConfigParam(ctx context.Context, key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	bypass := ContextWithBypass(ctx, true)
	existing, err := Search(bypass, configParamModel, [][]interface{}{{"key", "=", key}})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		id, _ := CoerceInt64(existing[0]["id"])
		return UpdateRecordByID(bypass, configParamModel, int(id), map[string]interface{}{"value": value})
	}
	m, ok := Registry[configParamModel]
	if !ok {
		return nil
	}
	_, err = Create(bypass, m, map[string]interface{}{"key": key, "value": value})
	return err
}

// ConfigParamBool parses a config parameter as boolean (true/1/t/yes).
func ConfigParamBool(ctx context.Context, key string, defaultVal bool) bool {
	raw := strings.ToLower(GetConfigParam(ctx, key, ""))
	if raw == "" {
		return defaultVal
	}
	return raw == "true" || raw == "1" || raw == "t" || raw == "yes"
}

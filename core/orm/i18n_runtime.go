package orm

import (
	"context"
	"strings"
)

// TranslateFieldLabel resolves a field label for the user's language from sys.translation.
func TranslateFieldLabel(ctx context.Context, modelName, fieldName, defaultLabel string) string {
	if DB == nil {
		return defaultLabel
	}
	lang := UserLang(ctx)
	if lang == "" || lang == "en_US" {
		return defaultLabel
	}
	key := modelName + "," + fieldName
	rows, err := Search(ctx, "sys.translation", [][]interface{}{
		{"lang", "=", lang},
		{"src", "=", key},
	})
	if err != nil || len(rows) == 0 {
		return defaultLabel
	}
	if v := strings.TrimSpace(AsString(rows[0]["value"])); v != "" {
		return v
	}
	return defaultLabel
}

// UserLang returns the active language code for ctx, defaulting to en_US.
func UserLang(ctx context.Context) string {
	_ = ctx
	return "en_US"
}

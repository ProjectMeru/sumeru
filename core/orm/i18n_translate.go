package orm

import (
	"context"
	"database/sql"
	"strings"
)

// Translate returns the translated value for src in lang, or src if missing.
func Translate(ctx context.Context, lang, src string) string {
	src = strings.TrimSpace(src)
	lang = strings.TrimSpace(lang)
	if src == "" || lang == "" || lang == "en_US" || DB == nil {
		return src
	}
	if _, ok := Registry["sys.translation"]; !ok {
		return src
	}
	var val sql.NullString
	err := DB.QueryRowContext(AuditedBypass(ctx, "i18n.translate"),
		`SELECT value FROM `+MustQuotedTableName("sys.translation")+
			` WHERE lang = $1 AND src = $2 LIMIT 1`, lang, src,
	).Scan(&val)
	if err != nil || !val.Valid || val.String == "" {
		return src
	}
	return val.String
}

// TranslateFieldLabel resolves a field label for the user's language from sys.translation.
func TranslateFieldLabel(ctx context.Context, modelName, fieldName, defaultLabel string) string {
	lang := UserLang(ctx)
	if lang == "" || lang == "en_US" {
		return defaultLabel
	}
	key := modelName + "," + fieldName
	if v := Translate(ctx, lang, key); v != key {
		return v
	}
	return defaultLabel
}

// UserLang returns core.user.lang for the user on ctx, or en_US when unset.
func UserLang(ctx context.Context) string {
	if DB == nil {
		return "en_US"
	}
	uid := UIDFromContext(ctx)
	if uid <= 0 {
		return "en_US"
	}
	user, err := SearchOne(ctx, "core.user", map[string]interface{}{"id": uid})
	if err != nil {
		return "en_US"
	}
	if lang := strings.TrimSpace(AsString(user["lang"])); lang != "" {
		return lang
	}
	return "en_US"
}

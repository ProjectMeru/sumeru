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

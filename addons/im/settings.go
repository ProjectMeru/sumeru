package im

import (
	"context"
	"database/sql"
	"strings"

	"sumeru/core/orm"
)

// IMEnabled reports whether internal chat is enabled for the active company (default true).
func IMEnabled(ctx context.Context) bool {
	if orm.DB == nil {
		return true
	}
	companyID := orm.CompanyIDFromContext(ctx)
	if 0 >= companyID {
		companyID = orm.ActiveCompanyIDForUser(ctx, orm.SecurityUID(ctx))
	}
	if enabled, ok := readIMEnabledColumn(ctx, int(companyID)); ok {
		return enabled
	}
	if enabled, ok := readLegacyChatterColumn(ctx, int(companyID)); ok {
		return enabled
	}
	return true
}

func readIMEnabledColumn(ctx context.Context, companyID int) (enabled bool, ok bool) {
	tn := orm.MustQuotedTableName("core.company")
	q := `SELECT im_enabled FROM ` + tn
	var imEnabled sql.NullBool
	var err error
	if companyID > 0 {
		err = orm.DB.QueryRowContext(ctx, q+` WHERE id = $1`, companyID).Scan(&imEnabled)
	} else {
		err = orm.DB.QueryRowContext(ctx, q+` ORDER BY id ASC LIMIT 1`).Scan(&imEnabled)
	}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "im_enabled") {
			return false, false
		}
		return true, true
	}
	if imEnabled.Valid {
		return imEnabled.Bool, true
	}
	return true, true
}

func readLegacyChatterColumn(ctx context.Context, companyID int) (enabled bool, ok bool) {
	tn := orm.MustQuotedTableName("core.company")
	q := `SELECT mail_chatter_enabled FROM ` + tn
	var legacy sql.NullBool
	var err error
	if companyID > 0 {
		err = orm.DB.QueryRowContext(ctx, q+` WHERE id = $1`, companyID).Scan(&legacy)
	} else {
		err = orm.DB.QueryRowContext(ctx, q+` ORDER BY id ASC LIMIT 1`).Scan(&legacy)
	}
	if err != nil {
		return true, true
	}
	if legacy.Valid {
		return legacy.Bool, true
	}
	return true, true
}

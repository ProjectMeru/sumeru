package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// CheckModelAccess verifies uid may perform op on model (read|write|create|unlink).
func CheckModelAccess(ctx context.Context, uid int, model string, op string) error {
	if SecurityBypass(ctx) {
		return nil
	}
	if uid == superuserUID {
		return nil
	}
	if _, ok := Registry[model]; !ok {
		return fmt.Errorf("unknown model %q", model)
	}
	if uid <= 0 {
		LogAccessDeny(ctx, model, op, "not authenticated")
		return &AccessDeniedError{Model: model, Operation: op}
	}
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return err
	}
	want, err := QuotedPermColumnForOp(op)
	if err != nil {
		return fmt.Errorf("unknown operation %q", op)
	}
	accTbl := MustQuotedTableName("sys.access")
	q := `SELECT group_id, ` + want + ` FROM ` + accTbl + ` WHERE model = $1 AND ` + want + ` = true`
	rows, err := DB.QueryContext(ctx, q, model)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var gid sql.NullInt64
		var perm bool
		if err := rows.Scan(&gid, &perm); err != nil {
			return err
		}
		if !perm {
			continue
		}
		if !gid.Valid || gid.Int64 == 0 {
			return nil
		}
		if intSliceContains(groups, int(gid.Int64)) {
			return nil
		}
	}
	LogAccessDeny(ctx, model, op, "no matching ACL")
	return &AccessDeniedError{Model: model, Operation: op}
}

func permColumnForOp(op string) string {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "read":
		return "perm_read"
	case "write":
		return "perm_write"
	case "create":
		return "perm_create"
	case "unlink":
		return "perm_unlink"
	default:
		return ""
	}
}

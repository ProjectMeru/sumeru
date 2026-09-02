package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// CheckFieldWriteAccess errors if any key in values is write-denied by sys.field.access.
func CheckFieldWriteAccess(ctx context.Context, uid int, model string, values map[string]interface{}) error {
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	denied, err := fieldAccessDenied(ctx, uid, model, "write")
	if err != nil {
		return err
	}
	for k := range values {
		if denied[k] {
			return fmt.Errorf("field access denied: %s.%s", model, k)
		}
	}
	return nil
}

func fieldAccessDenied(ctx context.Context, uid int, model, op string) (map[string]bool, error) {
	out := map[string]bool{}
	if _, ok := Registry["sys.field.access"]; !ok || DB == nil {
		return applyGroupsFieldDenial(ctx, uid, model, op, out)
	}
	col, err := QuotedPermColumnForOp(op)
	if err != nil {
		return nil, err
	}
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	tbl := MustQuotedTableName("sys.field.access")
	rows, err := DB.QueryContext(ctx,
		`SELECT field_name, group_id, `+col+` FROM `+tbl+` WHERE model = $1`, model)
	if err != nil {
		return out, nil
	}
	defer rows.Close()
	type rule struct {
		gid  int
		perm bool
	}
	byField := map[string][]rule{}
	for rows.Next() {
		var field string
		var gid sql.NullInt64
		var perm bool
		if err := rows.Scan(&field, &gid, &perm); err != nil {
			return nil, err
		}
		g := 0
		if gid.Valid {
			g = int(gid.Int64)
		}
		byField[field] = append(byField[field], rule{gid: g, perm: perm})
	}
	for field, rules := range byField {
		allowed := false
		matched := false
		for _, r := range rules {
			if r.gid == 0 || intSliceContains(groups, r.gid) {
				matched = true
				if r.perm {
					allowed = true
					break
				}
			}
		}
		if matched && !allowed {
			out[field] = true
		}
	}
	out, err = applyGroupsFieldDenial(ctx, uid, model, op, out)
	return out, rows.Err()
}

func applyGroupsFieldDenial(ctx context.Context, uid int, model, op string, out map[string]bool) (map[string]bool, error) {
	inst, ok := Registry[model]
	if !ok || inst == nil {
		return out, nil
	}
	for _, fieldDef := range inst.Fields() {
		if strings.TrimSpace(fieldDef.Groups) == "" {
			continue
		}
		if out[fieldDef.Name] {
			continue
		}
		if !UserHasGroupXML(ctx, uid, fieldDef.Groups) {
			out[fieldDef.Name] = true
		}
	}
	return out, nil
}

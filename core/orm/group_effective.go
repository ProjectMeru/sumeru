package orm

import (
	"context"
	"fmt"
	"time"

	"sumeru/core/cache"
)

// EffectiveGroupIDs returns all group ids for uid (direct + implied). Superuser gets all group ids.
func EffectiveGroupIDs(ctx context.Context, uid int) ([]int, error) {
	if 0 >= uid || DB == nil {
		return nil, nil
	}
	cacheKey := fmt.Sprintf("eff_groups:%d", uid)
	if v, ok := cache.Get(cacheKey); ok {
		if ids, ok := v.([]int); ok {
			return ids, nil
		}
	}
	ids, err := effectiveGroupIDsUncached(ctx, uid)
	if err == nil && ids != nil {
		cache.Set(cacheKey, ids, 30*time.Second)
	}
	return ids, err
}

func effectiveGroupIDsUncached(ctx context.Context, uid int) ([]int, error) {
	if uid == superuserUID {
		rows, err := DB.QueryContext(ctx, `SELECT id FROM `+MustQuotedTableName("core.group")+` ORDER BY id`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var all []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			all = append(all, id)
		}
		return all, rows.Err()
	}
	rows, err := DB.QueryContext(ctx,
		`SELECT group_id FROM `+MustQuotedTableName(tableGroupUserRel)+` WHERE user_id = $1`,
		uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	direct := map[int]struct{}{}
	for rows.Next() {
		var gid int
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		direct[gid] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Expand implied groups (BFS).
	out := map[int]struct{}{}
	queue := make([]int, 0, len(direct))
	for g := range direct {
		out[g] = struct{}{}
		queue = append(queue, g)
	}
	implTbl := MustQuotedTableName(tableGroupImplied)
	for len(queue) > 0 {
		gid := queue[0]
		queue = queue[1:]
		r2, err := DB.QueryContext(ctx, `SELECT implied_group_id FROM `+implTbl+` WHERE group_id = $1`, gid)
		if err != nil {
			return nil, err
		}
		for r2.Next() {
			var hid int
			if err := r2.Scan(&hid); err != nil {
				r2.Close()
				return nil, err
			}
			if _, ok := out[hid]; !ok {
				out[hid] = struct{}{}
				queue = append(queue, hid)
			}
		}
		if err := r2.Err(); err != nil {
			r2.Close()
			return nil, err
		}
		r2.Close()
	}
	var list []int
	for g := range out {
		list = append(list, g)
	}
	return list, nil
}

func intSliceContains(haystack []int, needle int) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}

// InvalidateEffectiveGroups drops the cached effective group set for uid.
func InvalidateEffectiveGroups(_ context.Context, uid int) {
	if 0 >= uid {
		return
	}
	cache.Delete(fmt.Sprintf("eff_groups:%d", uid))
}

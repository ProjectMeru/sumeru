package orm

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/cache"
)

type ruleDomainParts struct {
	globals        [][][]interface{}
	groups         [][][]interface{}
	allowAllGroups bool
}

var (
	ruleCacheMu sync.RWMutex
	ruleCache   = map[string]cachedRules{}
)

type cachedRules struct {
	parts ruleDomainParts
	until time.Time
}

func ruleCacheKey(uid int, model, op string, groups []int, companyIDs []int64) string {
	g := append([]int(nil), groups...)
	sort.Ints(g)
	c := append([]int64(nil), companyIDs...)
	sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
	return fmt.Sprintf("%d|%s|%s|%v|%v", uid, model, op, g, c)
}

// InvalidateRuleCache clears compiled record-rule caches (call after security XML/CSV sync).
func InvalidateRuleCache() {
	ruleCacheMu.Lock()
	ruleCache = map[string]cachedRules{}
	ruleCacheMu.Unlock()
	cache.DeletePrefix("eff_groups:")
	InvalidateXmlIDCache()
}

func loadRuleDomainParts(ctx context.Context, uid int, model string, op string) (ruleDomainParts, error) {
	var empty ruleDomainParts
	if SecurityBypass(ctx) || uid == superuserUID {
		return empty, nil
	}
	if uid <= 0 {
		return empty, nil
	}
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return empty, err
	}
	dc := DomainContext{UID: uid}
	if cids, err := UserCompanyIDs(ctx, uid); err == nil {
		dc.CompanyIDs = cids
		if len(cids) > 0 {
			dc.CompanyID = cids[0]
		}
	}
	key := ruleCacheKey(uid, model, op, groups, dc.CompanyIDs)
	ruleCacheMu.RLock()
	if c, ok := ruleCache[key]; ok && time.Now().Before(c.until) {
		ruleCacheMu.RUnlock()
		return c.parts, nil
	}
	ruleCacheMu.RUnlock()

	parts, err := loadRuleDomainsFromDB(ctx, model, op, groups, dc)
	if err != nil {
		return empty, err
	}
	ruleCacheMu.Lock()
	ruleCache[key] = cachedRules{parts: parts, until: time.Now().Add(30 * time.Second)}
	ruleCacheMu.Unlock()
	return parts, nil
}

func loadRuleDomainsFromDB(ctx context.Context, model, op string, groups []int, dc DomainContext) (ruleDomainParts, error) {
	var out ruleDomainParts
	permCol, err := QuotedPermColumnForOp(op)
	if err != nil {
		permCol, err = QuotedPermColumnForOp("read")
		if err != nil {
			return out, err
		}
	}
	ruleTbl := MustQuotedTableName("sys.rule")
	relTbl := MustQuotedTableName(tableRuleGroupRel)
	q := `SELECT r.id, r.domain_force, r.active, r.` + permCol + ` FROM ` + ruleTbl + ` r WHERE r.model = $1 AND r.active = true AND r.` + permCol + ` = true`
	rows, err := DB.QueryContext(ctx, q, model)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	type ruleRow struct {
		id          int
		domainForce string
	}
	var list []ruleRow
	for rows.Next() {
		var id int
		var domainForce string
		var active, permOp bool
		if err := rows.Scan(&id, &domainForce, &active, &permOp); err != nil {
			return out, err
		}
		if !active || !permOp {
			continue
		}
		list = append(list, ruleRow{id: id, domainForce: domainForce})
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	groupsByRule := map[int][]int{}
	if len(list) > 0 {
		idList := make([]interface{}, len(list))
		ph := make([]string, len(list))
		for i, r := range list {
			idList[i] = r.id
			ph[i] = fmt.Sprintf("$%d", i+1)
		}
		gq := `SELECT rule_id, group_id FROM ` + relTbl + ` WHERE rule_id IN (` + strings.Join(ph, ",") + `)`
		gr, err := DB.QueryContext(ctx, gq, idList...)
		if err != nil {
			return out, err
		}
		for gr.Next() {
			var rid, gid int
			if err := gr.Scan(&rid, &gid); err != nil {
				gr.Close()
				return out, err
			}
			groupsByRule[rid] = append(groupsByRule[rid], gid)
		}
		gr.Close()
	}
	for _, r := range list {
		dom, err := ParseDomainJSON(r.domainForce)
		if err != nil {
			return out, fmt.Errorf("rule %d: %w", r.id, err)
		}
		dom = SubstituteDomainContext(dom, dc)
		gids := groupsByRule[r.id]
		if len(gids) == 0 {
			if len(dom) > 0 {
				out.globals = append(out.globals, dom)
			}
			continue
		}
		match := false
		for _, gid := range gids {
			if intSliceContains(groups, gid) {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		if len(dom) == 0 {
			out.allowAllGroups = true
			continue
		}
		out.groups = append(out.groups, dom)
	}
	return out, nil
}

// BuildWhereWithRecordRules builds WHERE for user domain AND compiled record rules with correct parentheses.
func BuildWhereWithRecordRules(ctx context.Context, uid int, model, op string, base [][]interface{}) (string, []interface{}, error) {
	parts, err := loadRuleDomainParts(ctx, uid, model, op)
	if err != nil {
		return "", nil, err
	}
	if DevFeatureEnabled("access") {
		applog.L(ctx).Debug("dev_access_rules",
			"model", model, "op", op, "uid", uid,
			"global_rules", len(parts.globals), "group_rules", len(parts.groups),
			"allow_all_groups", parts.allowAllGroups,
		)
	}
	var clauses [][][]interface{}
	if len(base) > 0 {
		clauses = append(clauses, base)
	}
	clauses = append(clauses, parts.globals...)
	if !parts.allowAllGroups {
		switch len(parts.groups) {
		case 0:
		case 1:
			clauses = append(clauses, parts.groups[0])
		default:
			orSQL, orArgs, err := buildGroupORWhere(model, parts.groups)
			if err != nil {
				return "", nil, err
			}
			andSQL, andArgs, err := buildAndWhereClauses(model, clauses)
			if err != nil {
				return "", nil, err
			}
			if andSQL == "1=1" && len(clauses) == 0 {
				return orSQL, orArgs, nil
			}
			if len(clauses) == 0 {
				return orSQL, orArgs, nil
			}
			shifted, err := shiftPlaceholders(orSQL, len(andArgs)+1)
			if err != nil {
				return "", nil, err
			}
			return andSQL + " AND (" + shifted + ")", append(andArgs, orArgs...), nil
		}
	}
	return buildAndWhereClauses(model, clauses)
}

func buildGroupORWhere(modelName string, groups [][][]interface{}) (string, []interface{}, error) {
	return joinShiftedWhereFragments(" OR ", modelName, groups, 1)
}

// CheckRecordRules verify record satisfies all applicable rules.
func CheckRecordRules(ctx context.Context, uid int, model string, op string, rec map[string]interface{}) error {
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	if uid <= 0 {
		return fmt.Errorf("access denied")
	}
	parts, err := loadRuleDomainParts(ctx, uid, model, op)
	if err != nil {
		return err
	}
	for _, dom := range parts.globals {
		if !RecordMatchesDomain(rec, dom) {
			return fmt.Errorf("record rule failed for model %s", model)
		}
	}
	if parts.allowAllGroups {
		return nil
	}
	if len(parts.groups) == 0 {
		return nil
	}
	for _, dom := range parts.groups {
		if RecordMatchesDomain(rec, dom) {
			return nil
		}
	}
	return fmt.Errorf("record rule failed for model %s", model)
}

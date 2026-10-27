package module

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"sumeru/core/orm"
)

// impliedGroupLinkRefRE matches (4, ref('module.xml_id')) link tuples in core.group implied_ids eval strings.
var impliedGroupLinkRefRE = regexp.MustCompile(`\(\s*4\s*,\s*ref\s*\(\s*['"]([^'"]+)['"]\s*\)\s*\)`)

// ExtractImpliedGroupXMLRefs returns XML id refs from a core.group implied_ids eval.
func ExtractImpliedGroupXMLRefs(evalStr string) []string {
	evalStr = strings.TrimSpace(evalStr)
	if evalStr == "" {
		return nil
	}
	matches := impliedGroupLinkRefRE.FindAllStringSubmatch(evalStr, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := strings.TrimSpace(m[1])
		if ref != "" {
			out = append(out, ref)
		}
	}
	return out
}

// syncCoreGroupImpliedFromEval parses implied_ids eval and inserts rows into core.group.implied.
func syncCoreGroupImpliedFromEval(ctx context.Context, moduleName string, groupID int, evalStr string) error {
	if 0 >= groupID || strings.TrimSpace(evalStr) == "" {
		return nil
	}
	refs := ExtractImpliedGroupXMLRefs(evalStr)
	if 0 == len(refs) {
		return fmt.Errorf("no (4, ref('…')) implied link commands in %q", evalStr)
	}
	implTbl := orm.MustQuotedTableName("core.group.implied")
	for _, ref := range refs {
		impliedID, err := resolveXMLIDInModule(ctx, moduleName, ref)
		if err != nil {
			return fmt.Errorf("resolve implied ref %q: %w", ref, err)
		}
		if 0 >= impliedID {
			return fmt.Errorf("resolve implied ref %q: id is 0", ref)
		}
		if impliedID == groupID {
			continue
		}
		_, err = orm.DB.ExecContext(ctx,
			`INSERT INTO `+implTbl+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`,
			groupID, impliedID)
		if err != nil {
			return fmt.Errorf("implied edge %d -> %d: %w", groupID, impliedID, err)
		}
	}
	return nil
}

// syncSysRuleGroupsFromEval parses groups eval [(4, ref('…'))] into sys.rule.group.rel.
func syncSysRuleGroupsFromEval(ctx context.Context, moduleName string, ruleID int, evalStr string) error {
	if 0 >= ruleID || strings.TrimSpace(evalStr) == "" {
		return nil
	}
	refs := ExtractImpliedGroupXMLRefs(evalStr)
	if 0 == len(refs) {
		return fmt.Errorf("no (4, ref('…')) group link commands in %q", evalStr)
	}
	relTbl := orm.MustQuotedTableName("sys.rule.group.rel")
	if _, err := orm.DB.ExecContext(ctx, `DELETE FROM `+relTbl+` WHERE rule_id = $1`, ruleID); err != nil {
		return err
	}
	for _, ref := range refs {
		gid, err := resolveXMLIDInModule(ctx, moduleName, ref)
		if err != nil {
			return fmt.Errorf("resolve rule group ref %q: %w", ref, err)
		}
		if 0 >= gid {
			continue
		}
		_, err = orm.DB.ExecContext(ctx,
			`INSERT INTO `+relTbl+` (rule_id, group_id) VALUES ($1, $2) ON CONFLICT (rule_id, group_id) DO NOTHING`,
			ruleID, gid)
		if err != nil {
			return fmt.Errorf("rule %d group %d: %w", ruleID, gid, err)
		}
	}
	return nil
}

// EnsureSystemImpliesManagerGroup links base.group_system → module Manager groups
// so Settings admins inherit every installed app Manager (and thus User) ladder.
func EnsureSystemImpliesManagerGroup(ctx context.Context, moduleName, recordXMLID string, groupID int) error {
	if 0 >= groupID {
		return nil
	}
	idLower := strings.ToLower(strings.TrimSpace(recordXMLID))
	if !strings.Contains(idLower, "manager") {
		return nil
	}
	sysGID, _, err := orm.ResolveXmlId(ctx, "base.group_system")
	if err != nil || 0 >= sysGID {
		return err
	}
	if sysGID == groupID {
		return nil
	}
	implTbl := orm.MustQuotedTableName("core.group.implied")
	_, err = orm.DB.ExecContext(ctx,
		`INSERT INTO `+implTbl+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`,
		sysGID, groupID)
	if err != nil {
		return fmt.Errorf("system implies %s.%s: %w", moduleName, recordXMLID, err)
	}
	return nil
}

// ReconcileSystemManagerImpliedInstalled runs ReconcileSystemManagerImpliedForModule for every installed module (idempotent).
func ReconcileSystemManagerImpliedInstalled(ctx context.Context) {
	if orm.DB == nil {
		return
	}
	names, err := orm.InstalledModuleNames(ctx)
	if err != nil {
		syncWarn(ctx, "Warning: reconcile system→manager installed modules: %v", err)
		return
	}
	for name := range names {
		ReconcileSystemManagerImpliedForModule(ctx, name)
	}
}

// ReconcileSystemManagerImpliedForModule ensures base.group_system implies each Manager group declared in module XML.
func ReconcileSystemManagerImpliedForModule(ctx context.Context, moduleName string) {
	if orm.DB == nil {
		return
	}
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return
	}
	rows, err := orm.Search(ctx, "sys.model.data", [][]interface{}{
		{"module", "=", moduleName},
		{"model", "=", "core.group"},
	})
	if err != nil {
		return
	}
	for _, row := range rows {
		xmlID := strings.TrimSpace(orm.AsString(row["name"]))
		gid64, ok := orm.CoerceInt64(row["core_id"])
		if !ok || gid64 <= 0 || xmlID == "" {
			continue
		}
		if err := EnsureSystemImpliesManagerGroup(ctx, moduleName, xmlID, int(gid64)); err != nil {
			syncWarn(ctx, "Warning: reconcile system→manager %s.%s: %v", moduleName, xmlID, err)
		}
	}
}

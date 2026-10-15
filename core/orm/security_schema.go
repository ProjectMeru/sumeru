package orm

import (
	"fmt"
	"strings"
)

// Security-related join tables and indexes (M2M groups, companies, implied groups).
//
// Complements column sync in schema_sync.go; these tables are not full ORM models.

// EnsureSecurityJoinIndexes creates composite unique indexes for M2M-style tables.
func EnsureSecurityJoinIndexes() error {
	if DB == nil {
		return nil
	}
	joinTables := []string{
		`CREATE TABLE IF NOT EXISTS ` + MustQuotedTableName(tableGroupUserRel) + ` (user_id BIGINT NOT NULL, group_id BIGINT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ` + MustQuotedTableName(tableGroupImplied) + ` (group_id BIGINT NOT NULL, implied_group_id BIGINT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ` + MustQuotedTableName(tableRuleGroupRel) + ` (rule_id BIGINT NOT NULL, group_id BIGINT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ` + MustQuotedTableName(tableUserCompanyRel) + ` (user_id BIGINT NOT NULL, company_id BIGINT NOT NULL)`,
	}
	for _, q := range joinTables {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("create join table: %w", err)
		}
	}

	indexStatements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS core_group_users_rel_user_group_uq ON ` + MustQuotedTableName(tableGroupUserRel) + ` (user_id, group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS core_group_implied_gid_hid_uq ON ` + MustQuotedTableName(tableGroupImplied) + ` (group_id, implied_group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS sys_rule_group_rel_rule_group_uq ON ` + MustQuotedTableName(tableRuleGroupRel) + ` (rule_id, group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS core_user_company_rel_uq ON ` + MustQuotedTableName(tableUserCompanyRel) + ` (user_id, company_id)`,
	}
	for _, q := range indexStatements {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("%s: %w", q, err)
		}
	}
	return nil
}

// NullableGroupIDForAccess returns SQL argument for sys.access.group_id (NULL if 0).
func NullableGroupIDForAccess(groupID int) interface{} {
	if 0 >= groupID {
		return nil
	}
	return int64(groupID)
}

// NormalizeAccessGroupList splits comma-separated XML ids.
func NormalizeAccessGroupList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

//go:build integration

package integration_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
	"sumeru/core/sdk"
)

// TestRecordRuleDomainRestrictsRead ensures a global read rule filters search results.
func TestRecordRuleDomainRestrictsRead(t *testing.T) {
	elev := setupSuite(t)
	ruleModel, ok := orm.Registry["sys.rule"]
	if !ok {
		t.Fatal("sys.rule missing")
	}
	ruleID, err := orm.Create(elev, ruleModel, map[string]interface{}{
		"name":          "suite draft only",
		"model":         "test.suite.item",
		"domain_force":  `[["state","=","draft"]]`,
		"perm_read":     true,
		"perm_write":    true,
		"perm_create":   true,
		"perm_unlink":   true,
		"active":        true,
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}
	t.Cleanup(func() {
		_, _ = sdk.Unlink(elev, "sys.rule", []int{ruleID})
		orm.InvalidateRuleCache()
	})
	orm.InvalidateRuleCache()

	doneID, err := sdk.Create(elev, "test.suite.item", map[string]interface{}{
		"name": "rule-done", "kind": "k", "amount": 1, "state": "done",
	})
	if err != nil {
		t.Fatalf("create done: %v", err)
	}
	t.Cleanup(func() { _, _ = sdk.Unlink(elev, "test.suite.item", []int{doneID}) })

	grp, err := orm.SearchOne(elev, "core.group", map[string]interface{}{"xml_id": "base.group_user"})
	if err != nil {
		t.Fatalf("group: %v", err)
	}
	groupID64, _ := sdk.CoerceInt64(grp["id"])
	groupID := int(groupID64)
	login := "rule_user@test.local"
	uid, err := orm.Create(elev, orm.Registry["core.user"], map[string]interface{}{
		"login": login, "name": "Rule", "user_type": "internal",
	})
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	t.Cleanup(func() { _, _ = sdk.Unlink(elev, "core.user", []int{uid}) })
	rel := orm.MustQuotedTableName("core_group_user_rel")
	if _, err := orm.DB.Exec(`INSERT INTO `+rel+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, uid, groupID); err != nil {
		t.Fatalf("group rel: %v", err)
	}
	accessModel := orm.Registry["sys.access"]
	_, err = orm.Create(elev, accessModel, map[string]interface{}{
		"model": "test.suite.item", "group_id": groupID,
		"perm_read": true, "perm_write": true, "perm_create": true, "perm_unlink": true,
	})
	if err != nil {
		t.Fatalf("access: %v", err)
	}

	userCtx := orm.ContextWithUID(context.Background(), uid)
	row, err := sdk.SearchOne(userCtx, "test.suite.item", map[string]interface{}{"id": doneID})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if row != nil {
		t.Fatal("done item must be hidden by record rule")
	}
}

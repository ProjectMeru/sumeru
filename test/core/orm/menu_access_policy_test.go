package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestUserMayAccessMenuNavigation_superuserBypassesTaggedGroups(t *testing.T) {
	ctx := context.Background()
	const superuserUID = 1
	if !orm.UserMayAccessMenuNavigation(ctx, superuserUID, "crm.group_crm_user") {
		t.Fatal("superuser should pass navigation menu check")
	}
	if !orm.UserMayAccessMenu(ctx, superuserUID, "crm.group_crm_user") {
		t.Fatal("superuser should pass strict menu check")
	}
}

func TestDropExtraExclusiveGroups(t *testing.T) {
	got := orm.DropExtraExclusiveGroupsForTest([]int{1, 2, 3, 4}, map[int]int{1: 10, 2: 10, 4: 11})
	if len(got) != 3 || got[0] != 1 || got[1] != 3 || got[2] != 4 {
		t.Fatalf("exclusive: %v", got)
	}
}

func TestApplyAccessDomain(t *testing.T) {
	globals, groups, allowAll := orm.ApplyAccessDomainForTest(0, nil, [][]interface{}{{"id", "=", int64(1)}})
	if globals != 1 || groups != 0 || allowAll {
		t.Fatalf("global domain globals=%d groups=%d allow=%v", globals, groups, allowAll)
	}
	globals, groups, allowAll = orm.ApplyAccessDomainForTest(7, []int{7}, [][]interface{}{{"id", "=", int64(1)}})
	if globals != 0 || groups != 1 || allowAll {
		t.Fatalf("group domain globals=%d groups=%d allow=%v", globals, groups, allowAll)
	}
	globals, groups, allowAll = orm.ApplyAccessDomainForTest(7, nil, nil)
	if globals != 0 || groups != 0 || allowAll {
		t.Fatalf("empty domain must not allow all: globals=%d groups=%d allow=%v", globals, groups, allowAll)
	}
}

func TestMergeAccessGroupCSV(t *testing.T) {
	if got := orm.MergeAccessGroupCSV("base.group_user,base.group_system", "base.group_portal"); got != "base.group_portal" {
		t.Fatalf("replace: %q", got)
	}
	got := orm.MergeAccessGroupCSV("base.group_user,base.group_system", "-base.group_user,base.group_portal")
	if got != "base.group_system,base.group_portal" {
		t.Fatalf("remove: %q", got)
	}
}

func TestUserMayAccessMenu_strictRejectsTaggedWithoutGroup(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 99999)
	if orm.UserMayAccessMenu(ctx, 99999, "crm.group_crm_user") {
		t.Fatal("unknown uid without groups should not pass strict tagged menu check")
	}
}

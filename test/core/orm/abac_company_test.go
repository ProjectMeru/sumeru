package orm_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
)

func TestUserAllowedCompany_failClosed(t *testing.T) {
	ctx := context.Background()
	if orm.UserAllowedCompany(ctx, 2, 0) {
		t.Fatal("cid 0 should be denied")
	}
	if orm.UserAllowedCompany(ctx, 2, 99) {
		t.Fatal("unknown company should be denied without DB membership")
	}
	if !orm.UserAllowedCompany(ctx, 1, 99) {
		t.Fatal("superuser should allow any company")
	}
}

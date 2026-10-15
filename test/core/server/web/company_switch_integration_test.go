//go:build integration

package web_test

import (
	"context"
	"testing"

	"sumeru/core/orm"
	"sumeru/core/server/web"
)

func TestSwitchActiveCompanyRejectsDisallowedCompany(t *testing.T) {
	integrationDB(t)
	userID := insertTestUser(t)
	ctx := orm.ContextWithUID(context.Background(), userID)

	companyTable := orm.MustQuotedTableName("core.company")
	userTable := orm.MustQuotedTableName("core.user")
	var allowedID, otherID int
	if err := orm.DB.QueryRow(`INSERT INTO `+companyTable+` (name) VALUES ('Allowed Co') RETURNING id`).Scan(&allowedID); err != nil {
		t.Fatalf("insert allowed company: %v", err)
	}
	t.Cleanup(func() { _, _ = orm.DB.Exec(`DELETE FROM `+companyTable+` WHERE id = $1`, allowedID) })
	if err := orm.DB.QueryRow(`INSERT INTO `+companyTable+` (name) VALUES ('Other Co') RETURNING id`).Scan(&otherID); err != nil {
		t.Fatalf("insert other company: %v", err)
	}
	t.Cleanup(func() { _, _ = orm.DB.Exec(`DELETE FROM `+companyTable+` WHERE id = $1`, otherID) })

	if _, err := orm.DB.Exec(`UPDATE `+userTable+` SET company_id = $1 WHERE id = $2`, allowedID, userID); err != nil {
		t.Fatalf("set user company: %v", err)
	}
	relTable := orm.MustQuotedTableName("core.user.company.rel")
	if _, err := orm.DB.Exec(`INSERT INTO `+relTable+` (user_id, company_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, allowedID); err != nil {
		t.Fatalf("link allowed company: %v", err)
	}

	web.SwitchActiveCompanyForTest(ctx, userID, otherID)

	var active int
	if err := orm.DB.QueryRow(`SELECT company_id FROM `+userTable+` WHERE id = $1`, userID).Scan(&active); err != nil {
		t.Fatalf("read company_id: %v", err)
	}
	if active != allowedID {
		t.Fatalf("company_id = %d, want %d (disallowed switch must not apply)", active, allowedID)
	}
}

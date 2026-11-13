//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	_ "sumeru/addons/base"

	"sumeru/core/modelreg"
	"sumeru/core/orm"
	"sumeru/core/sdk"
)

type secureItem struct {
	sdk.Model `sumeru:"model=test.secure.item"`
	Name      sdk.String               `sumeru:"string=Name"`
	CompanyID sdk.Many2One[sdk.Any]    `sumeru:"index,string=Company,comodel=core.company"`
}

func init() {
	modelreg.MustRegister("securitycompany", &secureItem{})
}

func securityIntegrationDB(t *testing.T) context.Context {
	t.Helper()
	dsn := os.Getenv("SUMERU_TEST_DSN")
	if dsn == "" {
		t.Skip("SUMERU_TEST_DSN not set")
	}
	preflight, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := preflight.Ping(); err != nil {
		_ = preflight.Close()
		t.Fatalf("db ping: %v", err)
	}
	_ = preflight.Close()
	orm.InitDB(dsn)
	if err := modelreg.ActivateAll(nil); err != nil {
		t.Fatal(err)
	}
	if !orm.IsInitialized() {
		t.Skip("database schema not bootstrapped (run sumeru -i base)")
	}
	if err := orm.SyncRegistrySchemaForNames([]string{"test.secure.item"}); err != nil {
		t.Fatalf("schema sync: %v", err)
	}
	return orm.ContextWithBypass(context.Background(), true)
}

func grantSecureItemAccess(t *testing.T, elevated context.Context, groupID int) {
	t.Helper()
	accessModel := orm.Registry["sys.access"]
	_, err := orm.Create(elevated, accessModel, map[string]interface{}{
		"model": "test.secure.item", "group_id": groupID,
		"perm_read": true, "perm_write": true, "perm_create": true, "perm_unlink": true,
	})
	if err != nil {
		t.Fatalf("create sys.access: %v", err)
	}
}

func createCompanyUser(t *testing.T, elevated context.Context, companyID int) int {
	t.Helper()
	userModel := orm.Registry["core.user"]
	grp, err := orm.SearchOne(elevated, "core.group", map[string]interface{}{"xml_id": "base.group_user"})
	if err != nil {
		t.Fatalf("search group_user: %v", err)
	}
	groupID64, _ := sdk.CoerceInt64(grp["id"])
	groupID := int(groupID64)
	login := fmt.Sprintf("sec_co_%d_%d@test.local", companyID, time.Now().UnixNano())
	uid, err := orm.Create(elevated, userModel, map[string]interface{}{
		"login": login, "name": "Co test", "company_id": companyID, "user_type": "internal",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = sdk.Unlink(elevated, "core.user", []int{uid})
	})
	rel := orm.MustQuotedTableName("core_group_user_rel")
	if _, err := orm.DB.Exec(`INSERT INTO `+rel+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, uid, groupID); err != nil {
		t.Fatalf("link group: %v", err)
	}
	userRel := orm.MustQuotedTableName("core_user_company_rel")
	if _, err := orm.DB.Exec(`INSERT INTO `+userRel+` (user_id, company_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, uid, companyID); err != nil {
		t.Fatalf("link company: %v", err)
	}
	grantSecureItemAccess(t, elevated, groupID)
	return uid
}

func userCtx(t *testing.T, uid int, companyID int64) context.Context {
	t.Helper()
	ctx := orm.ContextWithUID(context.Background(), uid)
	return orm.ContextWithCompanyID(ctx, companyID)
}

func TestCrossCompanyReadDenied(t *testing.T) {
	elev := securityIntegrationDB(t)
	companyTable := orm.MustQuotedTableName("core.company")
	var companyA, companyB int
	if err := orm.DB.QueryRow(`INSERT INTO ` + companyTable + ` (name) VALUES ('SecCo A') RETURNING id`).Scan(&companyA); err != nil {
		t.Fatalf("company A: %v", err)
	}
	t.Cleanup(func() { _, _ = orm.DB.Exec(`DELETE FROM ` + companyTable + ` WHERE id = $1`, companyA) })
	if err := orm.DB.QueryRow(`INSERT INTO ` + companyTable + ` (name) VALUES ('SecCo B') RETURNING id`).Scan(&companyB); err != nil {
		t.Fatalf("company B: %v", err)
	}
	t.Cleanup(func() { _, _ = orm.DB.Exec(`DELETE FROM ` + companyTable + ` WHERE id = $1`, companyB) })

	itemModel := orm.Registry["test.secure.item"]
	idB, err := orm.Create(elev, itemModel, map[string]interface{}{"name": "B-item", "company_id": companyB})
	if err != nil {
		t.Fatalf("create B: %v", err)
	}
	t.Cleanup(func() { _, _ = sdk.Unlink(elev, "test.secure.item", []int{idB}) })

	uidA := createCompanyUser(t, elev, companyA)
	ctxA := userCtx(t, uidA, int64(companyA))

	row, err := sdk.SearchOne(ctxA, "test.secure.item", map[string]interface{}{"id": idB})
	if err != nil {
		t.Fatalf("search_one: %v", err)
	}
	if row != nil {
		t.Fatal("cross-company row must not be visible")
	}
}

func TestCrossCompanyWriteDenied(t *testing.T) {
	elev := securityIntegrationDB(t)
	companyTable := orm.MustQuotedTableName("core.company")
	var companyA, companyB int
	if err := orm.DB.QueryRow(`INSERT INTO ` + companyTable + ` (name) VALUES ('WrCo A') RETURNING id`).Scan(&companyA); err != nil {
		t.Fatalf("company A: %v", err)
	}
	t.Cleanup(func() { _, _ = orm.DB.Exec(`DELETE FROM ` + companyTable + ` WHERE id = $1`, companyA) })
	if err := orm.DB.QueryRow(`INSERT INTO ` + companyTable + ` (name) VALUES ('WrCo B') RETURNING id`).Scan(&companyB); err != nil {
		t.Fatalf("company B: %v", err)
	}
	t.Cleanup(func() { _, _ = orm.DB.Exec(`DELETE FROM ` + companyTable + ` WHERE id = $1`, companyB) })

	itemModel := orm.Registry["test.secure.item"]
	idB, err := orm.Create(elev, itemModel, map[string]interface{}{"name": "write-B", "company_id": companyB})
	if err != nil {
		t.Fatalf("create B: %v", err)
	}
	t.Cleanup(func() { _, _ = sdk.Unlink(elev, "test.secure.item", []int{idB}) })

	uidA := createCompanyUser(t, elev, companyA)
	ctxA := userCtx(t, uidA, int64(companyA))

	_, err = sdk.Write(ctxA, "test.secure.item", []int{idB}, map[string]interface{}{"name": "hacked"})
	if err == nil {
		t.Fatal("expected cross-company write to fail")
	}
	_, err = sdk.Unlink(ctxA, "test.secure.item", []int{idB})
	if err == nil {
		t.Fatal("expected cross-company unlink to fail")
	}
}

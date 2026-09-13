//go:build integration

package mail_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"sumeru/addons/mail"
	"sumeru/core/orm"
)

const mailACLDomainForce = `["|", ["company_id", "=", false], ["company_id", "in", "$company_ids"]]`

func mailIntegrationCtx() context.Context {
	ctx := orm.ContextWithBypass(context.Background(), true)
	return orm.ContextWithUID(ctx, 1)
}

func initMailIntegrationDB(t *testing.T) context.Context {
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

	orm.InitDBWithPool(dsn, orm.DBPoolSettings{MaxOpenConns: 5, MaxIdleConns: 2})
	if !orm.IsInitialized() {
		t.Skip("database schema not bootstrapped (run sumeru -i base)")
	}
	if _, ok := orm.Registry["mail.message"]; !ok {
		t.Skip("mail.message not registered (install mail module)")
	}
	ctx := mailIntegrationCtx()
	ensureMailSecurity(t, ctx)
	return ctx
}

func ensureMailSecurity(t *testing.T, ctx context.Context) {
	t.Helper()
	userGID, _, err := orm.ResolveXmlId(ctx, "base.group_user")
	if err != nil || userGID <= 0 {
		t.Fatalf("resolve base.group_user: %v", err)
	}
	systemGID, _, err := orm.ResolveXmlId(ctx, "base.group_system")
	if err != nil || systemGID <= 0 {
		t.Fatalf("resolve base.group_system: %v", err)
	}
	portalGID, _, err := orm.ResolveXmlId(ctx, "base.group_portal")
	if err != nil || portalGID <= 0 {
		t.Fatalf("resolve base.group_portal: %v", err)
	}
	_ = portalGID

	accessModel := orm.RegistryModel("sys.access")
	upsertAccess := func(name, model string, groupID int, read, write, create, unlink bool) {
		t.Helper()
		_, err := orm.Upsert(ctx, accessModel, map[string]interface{}{
			"name":        name,
			"model":       model,
			"group_id":    groupID,
			"perm_read":   read,
			"perm_write":  write,
			"perm_create": create,
			"perm_unlink": unlink,
		}, "name")
		if err != nil {
			t.Fatalf("upsert access %s: %v", name, err)
		}
	}
	upsertAccess("access_mail_message_user", "mail.message", userGID, true, true, true, false)
	upsertAccess("access_mail_activity_user", "mail.activity", userGID, true, true, true, true)
	upsertAccess("access_mail_activity_type_user", "mail.activity.type", systemGID, true, true, true, true)
	upsertAccess("access_mail_activity_plan_user", "mail.activity.plan", systemGID, true, true, true, true)
	upsertAccess("access_mail_activity_plan_template_user", "mail.activity.plan.template", systemGID, true, true, true, true)
	upsertAccess("access_mail_template_user", "mail.template", systemGID, true, true, true, true)

	ruleModel := orm.RegistryModel("sys.rule")
	upsertRule := func(name, model string, unlink bool) int {
		t.Helper()
		id, err := orm.Upsert(ctx, ruleModel, map[string]interface{}{
			"name":         name,
			"model":        model,
			"domain_force": mailACLDomainForce,
			"active":       true,
			"perm_read":    true,
			"perm_write":   true,
			"perm_create":  true,
			"perm_unlink":  unlink,
		}, "name")
		if err != nil {
			t.Fatalf("upsert rule %s: %v", name, err)
		}
		linkRuleGroup(t, ctx, id, userGID)
		return id
	}
	upsertRule("Mail Message: multi-company", "mail.message", false)
	upsertRule("Mail Activity: multi-company", "mail.activity", true)
	orm.InvalidateRuleCache()
}

func linkRuleGroup(t *testing.T, ctx context.Context, ruleID, groupID int) {
	t.Helper()
	relTbl := orm.MustQuotedTableName("sys.rule.group.rel")
	_, err := orm.DB.ExecContext(ctx,
		`DELETE FROM `+relTbl+` WHERE rule_id = $1`, ruleID)
	if err != nil {
		t.Fatalf("clear rule groups: %v", err)
	}
	_, err = orm.DB.ExecContext(ctx,
		`INSERT INTO `+relTbl+` (rule_id, group_id) VALUES ($1, $2) ON CONFLICT (rule_id, group_id) DO NOTHING`,
		ruleID, groupID)
	if err != nil {
		t.Fatalf("link rule group: %v", err)
	}
}

func createMailTestCompany(t *testing.T, ctx context.Context, label string) int {
	t.Helper()
	model, ok := orm.Registry["core.company"]
	if !ok {
		t.Fatal("core.company not registered")
	}
	name := fmt.Sprintf("Mail ACL %s %d", label, time.Now().UnixNano())
	id, err := orm.Create(ctx, model, map[string]interface{}{"name": name})
	if err != nil {
		t.Fatalf("create company: %v", err)
	}
	return id
}

func createMailTestUser(t *testing.T, ctx context.Context, label string, groupIDs, companyIDs []int) int {
	t.Helper()
	userModel, ok := orm.Registry["core.user"]
	if !ok {
		t.Fatal("core.user not registered")
	}
	login := fmt.Sprintf("mail_acl_%s_%d", label, time.Now().UnixNano())
	userID, err := orm.Create(ctx, userModel, map[string]interface{}{
		"login": login,
		"name":  "Mail ACL " + label,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := orm.SetUserGroupLinks(ctx, userID, groupIDs); err != nil {
		t.Fatalf("set groups: %v", err)
	}
	if err := orm.SetUserCompanyLinks(ctx, userID, companyIDs); err != nil {
		t.Fatalf("set companies: %v", err)
	}
	return userID
}

func createMailMessage(t *testing.T, ctx context.Context, resModel string, resID, companyID int, body string) int {
	t.Helper()
	model, ok := orm.Registry["mail.message"]
	if !ok {
		t.Fatal("mail.message not registered")
	}
	vals := map[string]interface{}{
		"model":       resModel,
		"core_id":     resID,
		"body":        body,
		"subtype":     mail.SubtypeComment,
		"author":      "Test",
		"create_date": time.Now().UTC(),
		"company_id":  companyID,
	}
	id, err := orm.Create(ctx, model, vals)
	if err != nil {
		t.Fatalf("create mail.message: %v", err)
	}
	return id
}

func userCtx(uid int) context.Context {
	return orm.ContextWithUID(context.Background(), uid)
}

func TestMailMessageCrossCompanySearchDenied(t *testing.T) {
	ctx := initMailIntegrationDB(t)
	userGID, _, err := orm.ResolveXmlId(ctx, "base.group_user")
	if err != nil || userGID <= 0 {
		t.Fatalf("resolve group_user: %v", err)
	}

	companyA := createMailTestCompany(t, ctx, "A")
	companyB := createMailTestCompany(t, ctx, "B")
	userA := createMailTestUser(t, ctx, "A", []int{userGID}, []int{companyA})
	_ = createMailTestUser(t, ctx, "B", []int{userGID}, []int{companyB})

	msgA := createMailMessage(t, ctx, "core.user", 1, companyA, "secret A")
	msgB := createMailMessage(t, ctx, "core.user", 1, companyB, "secret B")

	ctxA := userCtx(userA)
	rows, err := orm.Search(ctxA, "mail.message", nil)
	if err != nil {
		t.Fatalf("search as user A: %v", err)
	}
	for _, row := range rows {
		id, _ := row["id"].(int)
		if id == msgB {
			t.Fatalf("user A must not see company B message id=%d", msgB)
		}
	}
	foundA := false
	for _, row := range rows {
		id, _ := row["id"].(int)
		if id == msgA {
			foundA = true
		}
	}
	if !foundA {
		t.Fatalf("user A should see own company message id=%d", msgA)
	}

	count, err := orm.SearchCount(ctxA, "mail.message", nil)
	if err != nil {
		t.Fatalf("search_count: %v", err)
	}
	for _, row := range rows {
		id, _ := row["id"].(int)
		if id == msgB {
			t.Fatal("search_count path must not include company B message")
		}
	}
	if count < 1 {
		t.Fatal("expected at least one visible message for user A")
	}
}

func TestMailMessagePortalUserDenied(t *testing.T) {
	ctx := initMailIntegrationDB(t)
	portalGID, _, err := orm.ResolveXmlId(ctx, "base.group_portal")
	if err != nil || portalGID <= 0 {
		t.Fatalf("resolve portal group: %v", err)
	}
	portalUID := createMailTestUser(t, ctx, "portal", []int{portalGID}, nil)
	ctxPortal := userCtx(portalUID)
	err = orm.CheckModelAccess(ctxPortal, portalUID, "mail.message", "read")
	if err == nil {
		t.Fatal("portal user should not have mail.message read access")
	}
	if !orm.IsAccessDenied(err) {
		t.Fatalf("expected access denied, got %v", err)
	}
}

func TestListCommentsRespectsCompanyRule(t *testing.T) {
	ctx := initMailIntegrationDB(t)
	userGID, _, err := orm.ResolveXmlId(ctx, "base.group_user")
	if err != nil || userGID <= 0 {
		t.Fatalf("resolve group_user: %v", err)
	}

	companyA := createMailTestCompany(t, ctx, "listA")
	companyB := createMailTestCompany(t, ctx, "listB")
	userA := createMailTestUser(t, ctx, "listUserA", []int{userGID}, []int{companyA})

	recordID := createMailTestUser(t, ctx, "target", []int{userGID}, []int{companyA, companyB})
	_ = createMailMessage(t, ctx, "core.user", recordID, companyB, "cross-company comment")

	ctxA := userCtx(userA)
	rows, err := mail.ListCommentsForRecord(ctxA, "core.user", int64(recordID), 50)
	if err != nil {
		t.Fatalf("ListCommentsForRecord: %v", err)
	}
	for _, row := range rows {
		if row.Body == "cross-company comment" {
			t.Fatal("user A must not see company B comment via chatter list")
		}
	}
}

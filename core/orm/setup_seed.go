package orm

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"

	"golang.org/x/crypto/bcrypt"
)

func seedXmlID(ctx context.Context, module, xmlName, model string, coreID int) {
	_, _ = Upsert(ctx, RegistryModel("sys.model.data"), map[string]interface{}{
		"module":  module,
		"name":    xmlName,
		"model":   model,
		"core_id": coreID,
	}, "name")
}

func ensureDefaultKernelGroups(ctx context.Context) (adminGID int, userGID int, err error) {
	ctx = ContextWithBypass(ctx, true)
	if DB == nil {
		return 0, 0, nil
	}
	if err := EnsureSecurityJoinIndexes(); err != nil {
		return 0, 0, err
	}
	groupModel, ok := Registry["core.group"]
	if !ok || groupModel == nil {
		return 0, 0, fmt.Errorf("core.group not registered")
	}
	catModel, ok := Registry["sys.module.category"]
	if !ok || catModel == nil {
		return 0, 0, fmt.Errorf("sys.module.category not registered")
	}

	catAdminID, err := Upsert(ctx, catModel, map[string]interface{}{
		"name":     "Administration",
		"sequence": 1,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap sys.module.category Administration: %w", err)
	}
	seedXmlID(ctx, "base", "module_category_administration", "sys.module.category", catAdminID)
	catUserTypesID, err := Upsert(ctx, catModel, map[string]interface{}{
		"name":     "User types",
		"sequence": 2,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap sys.module.category User types: %w", err)
	}
	seedXmlID(ctx, "base", "module_category_user_types", "sys.module.category", catUserTypesID)

	adminGID, err = Upsert(ctx, groupModel, map[string]interface{}{
		"name":        "Administration / Settings",
		"category_id": catAdminID,
		"sequence":    1,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap core.group admin: %w", err)
	}
	seedXmlID(ctx, "base", "group_system", "core.group", adminGID)

	userGID, err = Upsert(ctx, groupModel, map[string]interface{}{
		"name":        "User types / Internal User",
		"category_id": catUserTypesID,
		"sequence":    10,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap core.group user: %w", err)
	}
	seedXmlID(ctx, "base", "group_user", "core.group", userGID)

	portalGID, err := Upsert(ctx, groupModel, map[string]interface{}{
		"name":        "User types / Portal",
		"category_id": catUserTypesID,
		"sequence":    20,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap core.group portal: %w", err)
	}
	seedXmlID(ctx, "base", "group_portal", "core.group", portalGID)

	publicGID, err := Upsert(ctx, groupModel, map[string]interface{}{
		"name":        "User types / Public",
		"category_id": catUserTypesID,
		"sequence":    30,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap core.group public: %w", err)
	}
	seedXmlID(ctx, "base", "group_public", "core.group", publicGID)

	_, _ = DB.ExecContext(ctx, `INSERT INTO `+MustQuotedTableName(tableGroupImplied)+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`, adminGID, userGID)
	return adminGID, userGID, nil
}

func ensureBootstrapSecurity(ctx context.Context, first *SetupAdminParams) error {
	ctx = ContextWithBypass(ctx, true)
	if DB == nil {
		return nil
	}
	adminGID, userGID, err := ensureDefaultKernelGroups(ctx)
	if err != nil {
		return err
	}

	userModel, ok := Registry["core.user"]
	if !ok || userModel == nil {
		return fmt.Errorf("core.user not registered")
	}
	companyModel, ok := Registry["core.company"]
	if !ok || companyModel == nil {
		return fmt.Errorf("core.company not registered")
	}

	userTbl := MustQuotedTableName("core.user")
	var userCount int
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+userTbl).Scan(&userCount); err != nil {
		return fmt.Errorf("count users: %w", err)
	}

	if userCount > 0 {
		ensureBootstrapACLs(ctx, adminGID, userGID)
		ensurePlatformDefaults(ctx)
		return nil
	}

	if first == nil {
		return fmt.Errorf("no users in this database: open /setup in your browser to create the administrator and finish initialization")
	}

	compID, err := Upsert(ctx, companyModel, map[string]interface{}{
		"name": first.CompanyName,
	}, "name")
	if err != nil {
		return fmt.Errorf("bootstrap company: %w", err)
	}
	seedXmlID(ctx, "base", "main_company", "core.company", compID)

	login := strings.ToLower(first.Email)
	adminUID, err := Upsert(ctx, userModel, map[string]interface{}{
		"login":     login,
		"name":      first.FullName,
		"active":    true,
		"email":     first.Email,
		"lang":      first.Lang,
		"password":  "",
		"user_type": "internal",
	}, "login")
	if err != nil {
		return fmt.Errorf("bootstrap administrator: %w", err)
	}
	if adminUID == 0 {
		return fmt.Errorf("bootstrap administrator user id")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(first.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := DB.ExecContext(ctx, `UPDATE `+userTbl+` SET password = $1 WHERE id = $2`, string(hash), adminUID); err != nil {
		return fmt.Errorf("set administrator password: %w", err)
	}

	seedXmlID(ctx, "base", "user_admin", "core.user", adminUID)

	if _, err := DB.ExecContext(ctx, `INSERT INTO `+MustQuotedTableName(tableGroupUserRel)+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, adminUID, adminGID); err != nil {
		return err
	}
	if _, err := DB.ExecContext(ctx, `INSERT INTO `+MustQuotedTableName(tableGroupUserRel)+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, adminUID, userGID); err != nil {
		return err
	}
	_, _ = DB.ExecContext(ctx, `UPDATE `+userTbl+` SET company_id = $1 WHERE id = $2`, compID, adminUID)

	ensureBootstrapACLs(ctx, adminGID, userGID)
	ensurePlatformDefaults(ctx)
	return nil
}

func ensurePlatformDefaults(ctx context.Context) {
	_ = SetConfig(ctx, "auth.password_min_length", "8")
	if _, ok := Registry["sys.sequence"]; !ok {
		return
	}
	if _, err := SearchOne(ctx, "sys.sequence", map[string]interface{}{"code": "core.user.apikey"}); err == nil {
		return
	}
	inst := Registry["sys.sequence"]
	_, _ = Create(ctx, inst, map[string]interface{}{
		"name":        "API Key",
		"code":        "core.user.apikey",
		"prefix":      "KEY/",
		"padding":     4,
		"number_next": 1,
		"active":      true,
	})
}

func ensureBootstrapACLs(ctx context.Context, adminGID, userGID int) {
	installed, err := InstalledModuleNames(ctx)
	if err != nil {
		applog.WarnMsg(ctx, "orm", "bootstrap", "Could not list installed modules for ACL bootstrap", err, nil)
		installed = nil
	}
	for modelName := range Registry {
		if len(installed) > 0 && !ShouldMaterializeModel(modelName, installed) {
			continue
		}
		accName := fmt.Sprintf("access_%s_admin", strings.ReplaceAll(modelName, ".", "_"))
		if _, err := Upsert(ctx, RegistryModel("sys.access"), map[string]interface{}{
			"name":        accName,
			"model":       modelName,
			"group_id":    NullableGroupIDForAccess(adminGID),
			"perm_read":   true,
			"perm_write":  true,
			"perm_create": true,
			"perm_unlink": true,
		}, "name"); err != nil {
			applog.WarnMsg(ctx, "orm", "bootstrap", "Bootstrap ACL upsert failed",
				err, map[string]interface{}{"model": modelName})
		}
	}

	globalReads := []string{"sys.model.data", "sys.menu", "sys.action.window", "sys.view", "sys.module", "sys.module.category", "core.country", "core.country.state", "core.city"}
	for _, m := range globalReads {
		accName := fmt.Sprintf("access_%s_global_read", strings.ReplaceAll(m, ".", "_"))
		_, _ = Upsert(ctx, RegistryModel("sys.access"), map[string]interface{}{
			"name":        accName,
			"model":       m,
			"group_id":    nil,
			"perm_read":   true,
			"perm_write":  false,
			"perm_create": false,
			"perm_unlink": false,
		}, "name")
	}

	for _, pair := range []struct{ model, name string }{
		{"core.company", "access_core_company_user"},
		{"core.user", "access_core_user_user"},
		{"core.group", "access_core_group_user_read"},
		{"sys.access", "access_sys_access_user_read"},
		{"sys.rule", "access_sys_rule_user_read"},
	} {
		_, _ = Upsert(ctx, RegistryModel("sys.access"), map[string]interface{}{
			"name":        pair.name,
			"model":       pair.model,
			"group_id":    NullableGroupIDForAccess(userGID),
			"perm_read":   true,
			"perm_write":  false,
			"perm_create": false,
			"perm_unlink": false,
		}, "name")
	}
}

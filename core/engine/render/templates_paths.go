package render

import "path/filepath"

// Relative paths under TemplatesPath (core/engine/templates).
const (
	TemplateShellBase = "shell/base.html"

	TemplatePartialMenuIcons  = "partials/menu_icons.html"
	TemplatePartialHomeHub    = "partials/home_hub.html"
	TemplatePartialAppsModule = "partials/apps_module.html"
	TemplatePartialLoginBrand = "partials/login_brand_aside.html"
	TemplatePartialAuthHead   = "partials/auth_head.html"

	TemplateAuthLogin = "auth/login.html"
	TemplateAuthTOTP  = "auth/totp_login.html"
	TemplateAuthSetup = "auth/setup.html"

	TemplatePortalHome  = "portal/portal_home.html"
	TemplatePortalShare = "portal/portal_share.html"

	TemplatePagesHome             = "pages/home_dashboard.html"
	TemplatePagesApps             = "pages/apps.html"
	TemplatePagesAppLogs          = "pages/app_logs.html"
	TemplatePagesSettingsHub      = "pages/settings_hub.html"
	TemplatePagesSettingsAccount  = "pages/settings_account.html"
	TemplatePagesSettingsAuth     = "pages/settings_auth.html"
	TemplatePagesSettingsMail     = "pages/settings_mail.html"
	TemplatePagesSettingsFieldACL = "pages/settings_field_acl.html"
	TemplatePagesSettingsModelACL = "pages/settings_model_acl.html"
	TemplatePagesImportWizard     = "pages/import_wizard.html"
)

// TemplatePath joins templatesDir with a relative template path.
func TemplatePath(templatesDir, rel string) string {
	return filepath.Join(templatesDir, rel)
}

// ShellLayoutTemplateFiles returns files parsed for base.html layout rendering.
func ShellLayoutTemplateFiles(templatesDir string) []string {
	return []string{
		TemplatePath(templatesDir, TemplateShellBase),
		TemplatePath(templatesDir, TemplatePartialMenuIcons),
	}
}

// ShellInnerTemplateFiles returns files parsed for a shell inner page (pages/*.html).
func ShellInnerTemplateFiles(templatesDir, pageRel string) []string {
	return []string{
		TemplatePath(templatesDir, pageRel),
		TemplatePath(templatesDir, TemplatePartialMenuIcons),
		TemplatePath(templatesDir, TemplatePartialHomeHub),
		TemplatePath(templatesDir, TemplatePartialAppsModule),
	}
}

// AuthTemplateFiles returns files parsed for standalone auth pages (login, TOTP, setup).
func AuthTemplateFiles(templatesDir, pageRel string) []string {
	return []string{
		TemplatePath(templatesDir, pageRel),
		TemplatePath(templatesDir, TemplatePartialLoginBrand),
		TemplatePath(templatesDir, TemplatePartialAuthHead),
	}
}

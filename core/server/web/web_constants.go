package web

import "sumeru/core/engine/render"

// Shared routes used across web handlers.
const (
	rootRoute              = "/"
	loginRoute             = "/web/login"
	logoutRoute            = "/web/logout"
	sessionTouchRoute      = "/web/session/touch"
	homeRoute              = "/web/home"
	portalRoutePrefix      = "/portal"
	portalHomeRoute        = "/portal/home"
	portalRecordRoute      = "/portal/record"
	portalShareRoutePrefix = "/portal/share/"
	companySwitchRoute     = "/web/company/switch"
	kanbanMoveRoute        = "/web/kanban/move"
	moduleActionRoute      = "/web/module/action"
	resetPasswordRoute     = "/web/action/reset_password"
	createAPIKeyRoute      = "/web/action/create_api_key"
	pinnedAppsRoute        = "/web/user/pinned-apps"
	settingsRoute          = "/web/settings"
	appLogsRoute           = "/web/settings/app-logs"
	apiHealthRoute         = "/api/health"
	apiReadyRoute          = "/api/ready"
	apiRPCRoute            = "/api/rpc"
	metricsRoute           = "/metrics"
	setupRoute             = "/setup"
	setupInitRoute         = "/setup/init"
)

// Setup wizard limits and templates.
const (
	maxSetupInitBodyBytes = 1 << 17 // 128 KiB
	setupTemplateFile     = render.TemplateAuthSetup
	setupOperation        = "setup"
	setupCompleteMessage  = "Setup complete — server is restarting…"
	setupTokenHeader      = "X-Setup-Token"
	setupRateLimitMax     = 5
	forwardedForHeader    = "X-Forwarded-For"
)

// Shell layout templates and defaults.
const (
	defaultPageTitle = "Sumeru"
)

// Pinned apps API form fields and limits.
const (
	pinnedAppsModulesField = "modules"
	maxPinnedAppsBodyBytes = 1 << 20
)

// JSON API routes, limits, and content types.
const (
	maxRPCBodyBytes       = 4 << 20
	jsonContentTypePrefix = "application/json"
	rpcMetricRequests     = "sumeru_rpc_requests_total"
	rpcMetricDuration     = "sumeru_rpc_duration_seconds"
)

// Login page identifiers and form fields.
const (
	loginTemplateFile       = render.TemplateAuthLogin
	totpLoginTemplateFile   = render.TemplateAuthTOTP
	totpLoginRoute          = "/web/login/totp"
	loginCSRFCookie         = "sumeru_login_csrf"
	setupCSRFCookie         = "sumeru_setup_csrf"
	loginNextCookie         = "sumeru_login_next"
	pendingMFACookie        = "sumeru_pending_mfa"
	trustedDeviceCookie     = "sumeru_trusted_device"
	loginLockoutMaxFailures = 5
	loginField              = "login"
	passwordField           = "password"
	nextField               = "next"
	invalidLoginMessage     = "Invalid login or password."
	resetPasswordMsg        = "reset_requested"
	oauthDeniedMsg          = "oauth_denied"
	authLocalDisabledMsg    = "auth_local_disabled"
	authLocalConfigKey      = "auth.local_enabled"
	resetUserIDField        = "id"
)

// Auth HTTP headers.
const (
	apiKeyHeader     = "X-API-Key"
	authHeader       = "Authorization"
	requestIDHeader  = "X-Request-ID"
	authBearerPrefix = "bearer "
)

// ORM model names referenced by web handlers.
const (
	sysMenuModel     = "sys.menu"
	coreCompanyModel = "core.company"
	coreUserModel    = "core.user"
)

// ACL group XML ids.
const groupSystemXML = "base.group_system"

// Common HTTP response messages.
const (
	forbiddenMessage        = "Forbidden"
	invalidCSRFMessage      = "Invalid CSRF token"
	methodNotAllowedMessage = "Method not allowed"
	invalidFormMessage      = "Invalid form"
	unknownModelMessage     = "Unknown model"
)

// Flash query parameter appended to redirects after form actions.
const flashMessageParam = "msg"

const saveOKCreatedMsg = "save_ok_created"
const saveOKUpdatedMsg = "save_ok_updated"
const stageUpdatedMsg = "stage_updated"

// Structured web log fields shared by WebLogEvent helpers.
const (
	webLogComponent     = "web"
	webLogUnknownRoute  = "-"
	logStatusSuccess    = "success"
	logStatusFailure    = "failure"
	logStatusPartial    = "partial"
	logOperationRequest = "request"
)

// Home dashboard page identifiers.
const (
	homeMenuRootXMLID = "base.menu_home_root"
	homeInnerTemplate = render.TemplatePagesHome
	homePageTitle     = "Home"
	homeStylesheetURL = "/static/css/sumeru-home.css"
	homeEmptyMessage  = "No installed applications. Install apps from Apps."
	baseModuleName    = "base"
)

// App logs page identifiers.
const (
	appLogsMenuXMLID       = "base.menu_app_logs"
	appLogsInnerTemplate   = render.TemplatePagesAppLogs
	appLogsPageTitle       = "App Logs"
	appLogsBreadcrumb      = "Event Log"
	appLogModel            = "app.log"
	maxAppLogEvents        = 500
	workspaceStylesheetURL = "/static/css/sumeru-workspace.css"
	pagesStylesheetURL     = "/static/css/sumeru-pages.css"
)

// Settings hub page identifiers.
const (
	settingsRootMenuXMLID         = "base.menu_settings_root"
	settingsHubMenuXMLID          = render.MenuGeneralSettingsXMLID
	settingsHubInnerTemplate      = render.TemplatePagesSettingsHub
	settingsHubPageTitle          = "Settings"
	settingsHubStylesheetURL      = "/static/css/sumeru-settings-hub.css"
	settingsHubBodyClass          = " sum-body--settings-hub"
	settingsAccountInnerTemplate  = render.TemplatePagesSettingsAccount
	settingsFieldACLInnerTemplate = render.TemplatePagesSettingsFieldACL
	settingsModelACLInnerTemplate = render.TemplatePagesSettingsModelACL
	groupUserXML                  = "base.group_user"
)

// Apps module action form fields (POST apps_*).
const (
	appsLayoutField   = "apps_layout"
	appsFilterField   = "apps_filter"
	appsScopeField    = "apps_scope"
	appsSearchField   = "apps_q"
	appsCategoryField = "apps_category"
	appsGroupByField  = "apps_group_by"
)

// Apps module action form fields (POST do=, module=, etc.).
const (
	moduleActionField      = "do"
	moduleNameField        = "module"
	moduleRowIDField       = "module_row_id"
	moduleDisplayNameField = "display_name"
	moduleAuthorField      = "author"
	moduleDescriptionField = "description"
)

// Apps module lifecycle action names (POST do=).
const (
	moduleActionInstall    = "install"
	moduleActionUninstall  = "uninstall"
	moduleActionDeactivate = "deactivate"
	moduleActionActivate   = "activate"
	moduleActionSaveModule = "save_module"
)

// Module action flash message keys returned via Apps ?msg=.
const (
	moduleMsgMissingModule    = "missing_module"
	moduleMsgUnknownAction    = "unknown_action"
	moduleMsgSaved            = "saved"
	moduleMsgInvalidModuleRow = "invalid_module_row"
	moduleMsgModuleNotFound   = "module_not_found"
	moduleMsgModuleMismatch   = "module_mismatch"
)

// Company switch form field.
const companyIDFormField = "company_id"

// Workspace URL query parameters (/web?action=&menu_id=&view_type=&id=).
const (
	workspaceRoute            = render.WorkspaceRoute
	workspaceActionParam      = render.WorkspaceActionParam
	workspaceMenuIDParam      = render.WorkspaceMenuIDParam
	workspaceViewTypeParam    = render.WorkspaceViewTypeParam
	workspaceRecordIDParam    = render.WorkspaceRecordIDParam
	workspaceEditParam        = render.WorkspaceEditParam
	workspaceEditEnabledValue = "1"
	workspaceModelParam       = render.WorkspaceModelParam
	workspaceFilterParam      = render.WorkspaceFilterParam
	workspaceSortParam        = render.WorkspaceSortParam
	workspaceOffsetParam      = render.WorkspaceOffsetParam
	workspaceGroupByParam     = render.WorkspaceGroupByParam
	workspaceDomainParam      = render.WorkspaceDomainParam
)

// Workspace view modes and row limits.
const (
	workspaceViewModeList      = render.ViewModeList
	workspaceViewModeForm      = render.ViewModeForm
	workspaceViewModeKanban    = render.ViewModeKanban
	workspaceViewModePivot     = render.ViewModePivot
	workspaceViewModeGraph     = render.ViewModeGraph
	workspaceViewModeCalendar  = render.ViewModeCalendar
	workspaceViewModeGantt     = render.ViewModeGantt
	workspaceViewModeMap       = render.ViewModeMap
	workspaceViewModeCohort    = render.ViewModeCohort
	workspaceViewModeHierarchy = render.ViewModeHierarchy
	workspaceViewModeActivity  = render.ViewModeActivity
	workspaceViewModeIframe    = render.ViewModeIframe
	maxWorkspaceListRows       = 500
	maxWorkspaceKanbanRows     = 200
	workspaceListPageSize      = 40
)

// ORM models used by workspace handlers.
const (
	sysActionWindowModel   = "sys.action.window"
	sysActionURLModel      = "sys.action.url"
	resConfigSettingsModel = "res.config.settings"
	workspaceViewOpenOp    = "view_open"
)

// Workspace error message fragments mapped to HTTP status codes.
const (
	workspaceErrNoView    = "No view for model"
	workspaceErrNotFound  = "not found"
	workspaceErrInvalidID = "invalid id"
)

// Record save form fields excluded from ORM values.
const (
	passwordPlainField         = "password_plain"
	securityGroupIDsField      = "security_group_ids"
	securityGroupsTouchedField = "security_groups_touched"
	securityUserTypeField      = "security_user_type"
	companyIDsField            = "company_ids"
)

// CSV import route, limits, and multipart form fields.
const (
	importCSVRoute        = "/web/import/csv"
	exportCSVRoute        = "/web/export/csv"
	exportPDFRoute        = "/web/export/pdf"
	exportXLSXRoute       = "/web/export/xlsx"
	templatePDFRoute      = "/web/report/template-pdf"
	bulkTemplateRoute     = "/web/bulk/template"
	bulkUploadRoute       = "/web/bulk/upload"
	bulkConfirmRoute      = "/web/bulk/confirm"
	bulkCancelRoute       = "/web/bulk/cancel"
	bulkPreviewRoute      = "/web/bulk/preview"
	bulkDryRunRoute       = "/web/bulk/dry-run"
	bulkSaveTemplateRoute = "/web/bulk/save-template"
	exportAsyncRoute      = "/web/export/async"
	bulkStatusRoute       = "/web/bulk/status"
	importWizardRoute     = "/web/import"
	swcImportRoute        = "/web/swc/import"
	maxImportBodyBytes    = 8 << 20
	importModelField      = "model"
	importFileField       = "file"
	reportFieldsParam     = "fields"
	reportPageSizeParam   = "page_size"
	importModeField       = "import_mode"
	actionIDField         = "action"
)

const recordModelField = "model"

// Apps page routes, templates, and ORM model.
const (
	appsRoute         = "/web/apps"
	appsPageTitle     = "Apps"
	appsInnerTemplate = render.TemplatePagesApps
	appsModuleModel   = "sys.module"
)

// Apps browse query values (filter, scope, layout).
const (
	appsFilterAll         = "all"
	appsFilterInstalled   = "installed"
	appsFilterUninstalled = "uninstalled"
	appsScopeAll          = "all"
	appsScopeApps         = "apps"
	appsScopeTechnical    = "technical"
	appsLayoutGrid        = "grid"
	appsLayoutList        = "list"
	appsGroupByCategory   = "category"
	moduleStateInstalled  = "installed"
)

// View layout query parameter and legacy values.
const (
	layoutQueryParam   = "layout"
	legacyKanbanLayout = "kanban"
)

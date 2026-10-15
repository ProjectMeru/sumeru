package web

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sumeru/core/engine/render"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
	"sumeru/core/report"
	"sumeru/core/server/config"
)

// Type aliases for external tests in core/server/web (web_test package).
type (
	ShellPageOpts     = shellPageOpts
	SetupInitRequest  = setupInitRequest
	AppsBrowseState   = appsBrowseState
	ModuleRow         = moduleRow
	PinnedAppsRequest = pinnedAppsRequest
	WorkspaceRequest  = workspaceRequest
	AppsModule        = appsModule
	AppsModuleGroup   = appsModuleGroup
)

// Route and query constants for external tests.
const (
	TestHomeRoute              = homeRoute
	TestWorkspaceRoute         = workspaceRoute
	TestAppsRoute              = appsRoute
	TestSettingsRoute          = settingsRoute
	TestAppLogsRoute           = appLogsRoute
	TestRootRoute              = rootRoute
	TestSetupRoute             = setupRoute
	TestLoginRoute             = loginRoute
	TestLogoutRoute            = logoutRoute
	TestPinnedAppsRoute        = pinnedAppsRoute
	TestCompanySwitchRoute     = companySwitchRoute
	TestModuleActionRoute      = moduleActionRoute
	TestAPIRPCRoute            = apiRPCRoute
	TestAPIRPCRoutePrefix      = apiRPCRoutePrefix
	TestWorkspaceActionParam   = workspaceActionParam
	TestWorkspaceMenuIDParam   = workspaceMenuIDParam
	TestWorkspaceViewTypeParam = workspaceViewTypeParam
	TestWorkspaceRecordIDParam = workspaceRecordIDParam
	TestWorkspaceViewModeForm  = workspaceViewModeForm
	TestWorkspaceViewModeList  = workspaceViewModeList
	TestAppsLayoutList         = appsLayoutList
	TestAppsLayoutGrid         = appsLayoutGrid
	TestAppsGroupByCategory    = appsGroupByCategory
	TestAppsLayoutField        = appsLayoutField
	TestAppsCategoryField      = appsCategoryField
	TestAppsGroupByField       = appsGroupByField
	TestFlashMessageParam      = flashMessageParam
	TestOAuthDeniedMsg         = oauthDeniedMsg
	TestAuthLocalDisabledMsg   = authLocalDisabledMsg
	TestResetPasswordMsg       = resetPasswordMsg
	TestResetPasswordRoute     = resetPasswordRoute
	TestRecordModelField       = recordModelField
	TestNextField              = nextField
	TestModuleActionInstall    = moduleActionInstall
	TestLegacyKanbanLayout     = legacyKanbanLayout
	TestDefaultPageTitle       = defaultPageTitle
	TestSetupInitRoute         = setupInitRoute
	TestForwardedForHeader     = forwardedForHeader
	TestSetupTokenHeader       = setupTokenHeader
	TestModuleMsgUnknownAction = moduleMsgUnknownAction
	TestAppLogsPageTitle       = appLogsPageTitle
	TestAppLogsBreadcrumb      = appLogsBreadcrumb
	TestSettingsHubPageTitle   = settingsHubPageTitle
)

// Numeric and URL constants for external tests.
var (
	TestWorkspaceStylesheetURL         = workspaceStylesheetURL
	TestPagesStylesheetURL             = pagesStylesheetURL
	TestSettingsHubStylesheetURL       = settingsHubStylesheetURL
	TestSettingsHubScriptURL           = settingsHubScriptURL
	TestMaxRPCBodyBytes          int64 = maxRPCBodyBytes
	TestSetupRateLimitWindow           = setupRateLimitWindow
	TestSetupRateLimitMax              = setupRateLimitMax
)

func HTTPStatusFromWorkspaceError(err error) int { return httpStatusFromWorkspaceError(err) }

func WorkspaceQueryParams(r *http.Request) (actionQuery, menuQuery string) {
	return workspaceQueryParams(r)
}

func RedirectIfMenuAccessDenied(w http.ResponseWriter, r *http.Request, menuQuery string) bool {
	return redirectIfMenuAccessDenied(w, r, menuQuery)
}

func ParseMenuIDString(menuQuery string) (menuID int, ok bool) { return parseMenuIDString(menuQuery) }

func ResolveActionIDFromQuery(ctx context.Context, actionQuery string) int {
	return resolveActionIDFromQuery(ctx, actionQuery)
}

func MenuRecordActionID(menuRecord map[string]interface{}) (actionID int, ok bool) {
	return menuRecordActionID(menuRecord)
}

func ParseWorkspaceRequest(r *http.Request, actionID int) WorkspaceRequest {
	return parseWorkspaceRequest(r, actionID)
}

func WorkspaceRequestFields(req WorkspaceRequest) (actionID int, menuID, viewType, recordID string, formEdit bool) {
	return req.actionID, req.menuID, req.viewType, req.recordID, req.formEdit
}

func URLWithQueryParam(path, param, value string) (string, error) {
	return urlWithQueryParam(path, param, value)
}

func RedirectWithWebMessage(w http.ResponseWriter, r *http.Request, rawNext, message string) {
	redirectWithWebMessage(w, r, rawNext, message)
}

func NormalizeGridListLayout(raw string) string { return normalizeGridListLayout(raw) }

func LayoutFromQuery(r *http.Request) string { return layoutFromQuery(r) }

func LayoutFromForm(r *http.Request, fieldName string) string { return layoutFromForm(r, fieldName) }

func ParsePinnedAppsJSONBody(r *http.Request) (PinnedAppsRequest, bool) {
	return parsePinnedAppsJSONBody(r)
}

func ParsePinnedAppsFormBody(r *http.Request) (PinnedAppsRequest, bool) {
	return parsePinnedAppsFormBody(r)
}

func DecodePinnedModulesField(raw string) ([]string, bool) { return decodePinnedModulesField(raw) }

func NormalizePinnedModules(modules []string) []string { return normalizePinnedModules(modules) }

func ParsePinnedAppsRequest(r *http.Request) (PinnedAppsRequest, bool) {
	return parsePinnedAppsRequest(r)
}

func ResolveShellRoute(opts ShellPageOpts, r *http.Request) string { return resolveShellRoute(opts, r) }

func ResolveExtraStylesheets(pageStylesheets, optStylesheets []string) []string {
	return resolveExtraStylesheets(pageStylesheets, optStylesheets)
}

func ApplyShellPageDefaults(page render.PageData, opts ShellPageOpts, r *http.Request) render.PageData {
	return applyShellPageDefaults(page, opts, r)
}

func ClientIP(r *http.Request) string { return clientIP(r) }

func IsLoopbackIP(ip string) bool { return isLoopbackIP(ip) }

func SetupTokenFromRequest(r *http.Request, bodyToken string) string {
	return setupTokenFromRequest(r, bodyToken)
}

func PruneSetupAttempts(attempts []time.Time, now time.Time) []time.Time {
	return pruneAttemptsWithin(attempts, now, setupRateLimitWindow)
}

func AllowSetupRateLimit(w http.ResponseWriter, requestIP string) bool {
	return allowSetupRateLimit(w, requestIP)
}

func ValidateSetupToken(w http.ResponseWriter, r *http.Request, tokenFromBody string) bool {
	return validateSetupToken(w, r, tokenFromBody)
}

func RequireSetupEnvironmentForTest(w http.ResponseWriter, r *http.Request) bool {
	return requireSetupEnvironment(w, r)
}

func SetupClientIPForTest(r *http.Request) string {
	return setupClientIP(r)
}

func CheckSwcBusOrigin(r *http.Request) bool { return checkSwcBusOrigin(r) }

func ResetSetupRateLimiterForTest() { setupRateLimiter.attemptsByIP = make(map[string][]time.Time) }

func ParseSetupInitRequest(w http.ResponseWriter, body []byte) (SetupInitRequest, bool) {
	return parseSetupInitRequest(w, body)
}

func ToSetupAdminParams(request SetupInitRequest) orm.SetupAdminParams {
	return toSetupAdminParams(request)
}

func BuildSetupPageData() setupPageData { return buildSetupPageData() }

func BuildSettingsHubPageData(ctx context.Context, menuIDStr string) render.PageData {
	return buildSettingsHubPageData(ctx, menuIDStr)
}

func AppLogsViewStylesheets() []string { return appLogsViewStylesheets() }

func BuildAppLogsPageData(ctx context.Context, menuID int) render.PageData {
	return buildAppLogsPageData(ctx, menuID)
}

func ResolveMenuID(ctx context.Context, menuXMLID string) (menuID int, menuIDStr string) {
	return resolveMenuID(ctx, menuXMLID)
}

func AcceptsJSONContentType(contentType string) bool { return acceptsJSONContentType(contentType) }

func ParseRPCCallMeta(requestBody []byte) (modelName, methodName string) {
	return parseRPCCallMeta(requestBody)
}

func ReadBoundedRequestBody(r *http.Request, maxBytes int64) ([]byte, bool) {
	return readBoundedRequestBody(r, maxBytes)
}

func ServeMuxOrDefault(mux *http.ServeMux) *http.ServeMux { return serveMuxOrDefault(mux) }

func RootRedirectHome(w http.ResponseWriter, r *http.Request) { rootRedirectHome(w, r) }

func RootRedirectSetup(w http.ResponseWriter, r *http.Request) { rootRedirectSetup(w, r) }

func RedirectFoundTo(dest string) http.HandlerFunc { return redirectFoundTo(dest) }

func UserFacingRecordError(operation, model string, err error) (title, body, details string, fieldErrors []string) {
	return userFacingRecordError(operation, model, err)
}

func SetRecordErrorFlashForTest(w http.ResponseWriter, flash PageFlash) {
	SetRecordErrorFlash(w, flash)
}

func ConsumeRecordErrorFlashForTest(r *http.Request, w http.ResponseWriter) (PageFlash, bool) {
	return ConsumeRecordErrorFlash(r, w)
}

func ActionDefaultFieldValues(actionData map[string]interface{}) map[string]interface{} {
	return actionDefaultFieldValues(actionData)
}

func EnsureFormEditRedirectURL(rawNext string, clearRecordID bool) string {
	return ensureFormEditRedirectURL(rawNext, clearRecordID)
}

func SplitCommaSeparatedValues(raw string) []string { return splitCommaSeparatedValues(raw) }

func FirstGroupByField(raw string) string { return firstGroupByField(raw) }

func SplitViewModes(viewMode string) []string { return splitViewModes(viewMode) }

func NormalizeViewMode(viewMode string) string { return normalizeViewMode(viewMode) }

func FormBaseQueryValues(actionID int, menuID, viewType, recordID string) string {
	return formBaseQueryValues(actionID, menuID, viewType, recordID)
}

func WorkspaceListURL(actionID, menuID string) string { return workspaceListURL(actionID, menuID) }

func FormOrQueryValue(r *http.Request, field string) string { return formOrQueryValue(r, field) }

func CoerceCSVValue(raw string) interface{} { return coerceCSVValue(raw) }

func ImportableRowValues(header []string, record []string, allowedFields map[string]struct{}) map[string]interface{} {
	return importableRowValues(header, record, allowedFields)
}

func IsImportableColumn(column string, allowedFields map[string]struct{}) bool {
	return isImportableColumn(column, allowedFields)
}

func ImportCSVFlashMessage(createdCount int) string { return importCSVFlashMessage(createdCount) }

func ParseCompanySwitchForm(r *http.Request) companySwitchForm { return parseCompanySwitchForm(r) }

// SwitchActiveCompanyForTest exercises active company switching without HTTP.
func SwitchActiveCompanyForTest(ctx context.Context, userID, companyID int) {
	switchActiveCompany(ctx, userID, companyID)
}

// LoginURLWithReturn returns the login path (return target is stored in a cookie by redirectToLogin).
func LoginURLWithReturn(_ string) string { return loginRoute }

// SetLoginNextCookieForTest sets the pre-login return-path cookie for tests.
func SetLoginNextCookieForTest(w http.ResponseWriter, returnTo string) {
	setLoginNextCookie(w, returnTo)
}

// RedirectToLoginForTest mirrors requireLogin redirect without session check.
func RedirectToLoginForTest(w http.ResponseWriter, r *http.Request, returnTo string) {
	redirectToLogin(w, r, returnTo)
}

func BearerToken(header string) string { return bearerToken(header) }

func ParseLoginCredentials(r *http.Request) loginCredentials { return parseLoginCredentials(r) }

func AppsRedirectURL(msg string, browse AppsBrowseState) string { return appsRedirectURL(msg, browse) }

func ParseAppsBrowseStateFromForm(r *http.Request) AppsBrowseState {
	return parseAppsBrowseStateFromForm(r)
}

func AppsDetailRedirectURL(msg string, browse AppsBrowseState) string {
	return appsRedirectURL(msg, browse)
}

func ParseModuleActionForm(r *http.Request) moduleActionForm { return parseModuleActionForm(r) }

func RunModuleLifecycleAction(ctx context.Context, action, moduleName string) (string, error) {
	return runModuleLifecycleAction(ctx, action, moduleName)
}

func AppsModuleFromParsed(row ModuleRow) AppsModule { return appsModuleFromParsed(row) }

func EnrichAppsModules(ctx context.Context, modules []AppsModule, browse AppsBrowseState) []AppsModule {
	return enrichAppsModules(ctx, modules, browse)
}

func AppsFlashFromMessage(msg string, displayNames map[string]string) (render.FlashMessage, bool) {
	return appsFlashFromMessage(msg, displayNames)
}

func SplitAppsPageFlashes(msg string, displayNames map[string]string) (inline, toast []render.FlashMessage) {
	return splitAppsPageFlashes(msg, displayNames)
}

func FormatAppsActionError(message string) string { return formatAppsActionError(message) }

func ModuleStatusLabel(row ModuleRow) string { return moduleStatusLabel(row) }

func BuildModuleDisplayNameMap(modules []AppsModule) map[string]string {
	return buildModuleDisplayNameMap(modules)
}

func FilterAppsModulesByBrowse(modules []AppsModule, browse AppsBrowseState) (appModules, techModules []AppsModule) {
	return filterAppsModulesByBrowse(modules, browse)
}

func GroupAppsModules(modules []AppsModule, groupBy string) []AppsModuleGroup {
	return groupAppsModules(modules, groupBy)
}

func ModuleSummary(moduleName, description string) string {
	return moduleSummary(moduleName, description)
}

func ModuleSummaryFromDescription(description string) string {
	return moduleSummaryFromDescription(description)
}

func ModuleHasLongDescription(summary, description string) bool {
	return moduleHasLongDescription(summary, description)
}

func AppsLinkFromBrowse(browse AppsBrowseState) string {
	return appsLinkFromBrowse(browse)
}

func AppsDetailURL(browse AppsBrowseState, editing bool) string {
	return appsDetailPageURL(browse, editing)
}

func FindAppsModule(modules []AppsModule, moduleName string) (AppsModule, bool) {
	return findAppsModule(modules, moduleName)
}

func APIKeyTargetUserID(r *http.Request) int { return apiKeyTargetUserID(r) }

func APIKeyExpiresAtFromFormForTest(r *http.Request) (time.Time, string, error) {
	return apiKeyExpiresAtFromForm(r)
}

func ParseModuleRow(values map[string]interface{}) (ModuleRow, bool) { return parseModuleRow(values) }

func ModuleDisplayName(moduleName, displayName string) string {
	return moduleDisplayName(moduleName, displayName)
}

// ExportXLSXHandlerForTest exposes the XLSX export handler for external tests.
func ExportXLSXHandlerForTest(w http.ResponseWriter, r *http.Request) {
	ExportXLSXHandler(w, r)
}

// ResolveExportRequestForTest exposes export query validation for external tests.
func ResolveExportRequestForTest(w http.ResponseWriter, r *http.Request) (report.ExportCSVInput, bool) {
	return resolveExportRequest(w, r)
}

// ExportTemplatePDFHandlerForTest exposes the template PDF handler for external tests.
func ExportTemplatePDFHandlerForTest(w http.ResponseWriter, r *http.Request) {
	ExportTemplatePDFHandler(w, r)
}

func ContentHandlerForTest(w http.ResponseWriter, r *http.Request) {
	ContentHandler(w, r)
}

func ParseContentPathForTest(path string) (field string, id int, ok bool) {
	spec, ok := parseContentPath(contentRoutePrefix + path)
	if !ok {
		return "", 0, false
	}
	if spec.attachmentID > 0 {
		return "attachment", spec.attachmentID, true
	}
	return spec.resField, spec.resID, true
}

func ContentInlineAllowedForTest(mime string) bool { return contentInlineAllowed(mime) }

func SafeContentDispositionInlineForTest(name string) string {
	return safeContentDispositionInline(name)
}

func BulkPreviewHandlerForTest(w http.ResponseWriter, r *http.Request) {
	BulkPreviewHandler(w, r)
}

func BulkDryRunHandlerForTest(w http.ResponseWriter, r *http.Request) {
	BulkDryRunHandler(w, r)
}

func BulkSaveImportTemplateHandlerForTest(w http.ResponseWriter, r *http.Request) {
	BulkSaveImportTemplateHandler(w, r)
}

func BulkStatusHandlerForTest(w http.ResponseWriter, r *http.Request) {
	BulkStatusHandler(w, r)
}

// BulkBatchOwnedByForTest exposes bulk batch ownership checks for tests.
func BulkBatchOwnedByForTest(batch map[string]interface{}, uid int) bool {
	return bulkBatchOwnedBy(batch, uid)
}

func SwcImportHandlerForTest(w http.ResponseWriter, r *http.Request) {
	SwcImportHandler(w, r)
}

func SwcActivityLogHandlerForTest(w http.ResponseWriter, r *http.Request) {
	SwcActivityLogHandler(w, r)
}

func SwcDirectUsersHandlerForTest(w http.ResponseWriter, r *http.Request) {
	SwcDirectUsersHandler(w, r)
}

func SwcDirectThreadHandlerForTest(w http.ResponseWriter, r *http.Request) {
	SwcDirectThreadHandler(w, r)
}

func SwcDirectConversationsHandlerForTest(w http.ResponseWriter, r *http.Request) {
	SwcDirectConversationsHandler(w, r)
}

func SwcDirectPostHandlerForTest(w http.ResponseWriter, r *http.Request) {
	SwcDirectPostHandler(w, r)
}

func ApplyActivityRecordFlagsForTest(ctx context.Context, payload *swcmeta.WorkspacePayload, viewMode string, recordID int, model string) {
	applyActivityRecordFlags(ctx, payload, viewMode, recordID, model)
}

func ImportWizardPageHandlerForTest(w http.ResponseWriter, r *http.Request) {
	ImportWizardPageHandler(w, r)
}

func ExportAsyncHandlerForTest(w http.ResponseWriter, r *http.Request) {
	ExportAsyncHandler(w, r)
}

// SetTestSessionUserIDForTest overrides SessionUserID for handler tests.
func SetTestSessionUserIDForTest(userID int) { testSessionUserIDOverride = userID }

// ResetTestSessionUserIDForTest clears the session override.
func ResetTestSessionUserIDForTest() { testSessionUserIDOverride = 0 }

// SessionUserIDForTest exposes SessionUserID for external tests.
func SessionUserIDForTest(r *http.Request) int { return SessionUserID(r) }

// AuthViaSessionForTest exposes AuthViaSession for external tests.
func AuthViaSessionForTest(r *http.Request) bool { return AuthViaSession(r) }

// RPCJSONHandlerForTest exposes the JSON RPC handler for external tests.
func RPCJSONHandlerForTest(w http.ResponseWriter, r *http.Request) { RPCJSONHandler(w, r) }

// ApplyDefaultRPCHeadersForTest sets required RPC headers for handler tests.
func ApplyDefaultRPCHeadersForTest(r *http.Request) {
	if strings.TrimSpace(config.AppConfig.DbName) == "" {
		config.AppConfig.DbName = "test_db"
	}
	r.Header.Set(sumeruDatabaseHeader, config.AppConfig.DbName)
	if strings.TrimSpace(r.Header.Get(sumeruTenantHeader)) == "" {
		r.Header.Set(sumeruTenantHeader, "1")
	}
}

// TestRPCURLForTest builds a path-based RPC URL for tests.
func TestRPCURLForTest(model, method string) string {
	return apiRPCRoutePrefix + model + "/" + method
}

// TestAuthLoginURLForTest is POST /api/rpc/auth.login.
func TestAuthLoginURLForTest() string {
	return apiRPCRoutePrefix + rpcAuthLoginSegment
}

// BuildSessionCookieForTest exposes session cookie construction for tests.
func BuildSessionCookieForTest(value string, deleteCookie bool) *http.Cookie {
	return buildSessionCookie(value, deleteCookie)
}

// SessionCookieFromRequestForTest exposes session cookie lookup for tests.
func SessionCookieFromRequestForTest(r *http.Request) (sid, cookieName string) {
	return sessionCookieFromRequest(r)
}

// LoginGetForTest exposes the login page GET handler for tests.
func LoginGetForTest(w http.ResponseWriter, r *http.Request) { LoginGet(w, r) }

// OAuthCallbackForTest exposes the OAuth callback handler for tests.
func OAuthCallbackForTest(w http.ResponseWriter, r *http.Request) { OAuthCallbackHandler(w, r) }

// OAuthStartForTest exposes the OAuth start handler for tests.
func OAuthStartForTest(w http.ResponseWriter, r *http.Request) { OAuthStartHandler(w, r) }

// LinkOAuthUserForTest exposes OAuth account linking and JIT provisioning for tests.
func LinkOAuthUserForTest(ctx context.Context, providerID int, subject, email string, emailVerified bool, policy string) (int, error) {
	return linkOAuthUser(ctx, providerID, subject, email, emailVerified, policy)
}

// ResolveOAuthIdentityForTest exposes OIDC identity resolution for tests.
func ResolveOAuthIdentityForTest(ctx context.Context, prov map[string]interface{}, tok map[string]interface{}) (email, sub string, emailVerified bool, err error) {
	return resolveOAuthIdentity(ctx, prov, tok)
}

// PublicBaseURLForTest exposes canonical public origin resolution for tests.
func PublicBaseURLForTest(r *http.Request) (string, error) {
	return publicBaseURL(r)
}

// ValidateSetupCSRFForTest exposes setup wizard CSRF validation for tests.
func ValidateSetupCSRFForTest(r *http.Request, tokenFromBody string) bool {
	return validateSetupCSRF(r, tokenFromBody)
}

// TestSetupCSRFCookie is the HttpOnly setup CSRF cookie name.
const TestSetupCSRFCookie = setupCSRFCookie

// AuthLocalEnabledForTest reports whether password login is shown on the login page.
func AuthLocalEnabledForTest(ctx context.Context) bool { return authLocalEnabled(ctx) }

// LoginAuthProviderForTest is a login SSO button descriptor for tests.
type LoginAuthProviderForTest loginAuthProvider

// ListEnabledAuthProvidersForTest returns SSO providers for the login page.
func ListEnabledAuthProvidersForTest(ctx context.Context) []LoginAuthProviderForTest {
	rows := listEnabledAuthProviders(ctx)
	out := make([]LoginAuthProviderForTest, len(rows))
	for i, row := range rows {
		out[i] = LoginAuthProviderForTest(row)
	}
	return out
}

// LoginPageCompanyNameForTest returns branding text for the login page.
func LoginPageCompanyNameForTest(ctx context.Context) string { return loginPageCompanyName(ctx) }

// TestLoginBrandLogoRoute is the public company login logo URL.
const TestLoginBrandLogoRoute = loginBrandLogoRoute

// TestDefaultSumeruLogoURL is the bundled default sign-in logo path.
const TestDefaultSumeruLogoURL = defaultSumeruLogoURL

// LoginCompanyRowForTest configures synthetic login branding inputs.
type LoginCompanyRowForTest struct {
	Name         string
	LoginLogo    string
	LoginTagline string
	Color        int
}

// LoginBrandSnapshotForTest is the resolved left-panel branding for assertions.
type LoginBrandSnapshotForTest struct {
	LogoURL        string
	ShowPoweredBy  bool
	ShowTrustChips bool
}

// BuildLoginBrandForTest resolves branding from a synthetic company row (no DB).
func BuildLoginBrandForTest(row LoginCompanyRowForTest) LoginBrandSnapshotForTest {
	b := buildLoginBrand(loginCompanyRow(row))
	return LoginBrandSnapshotForTest{
		LogoURL:        b.LogoURL,
		ShowPoweredBy:  b.ShowPoweredBy,
		ShowTrustChips: b.ShowTrustChips,
	}
}

// ParseLoginLogoDataURLForTest decodes login logo data URLs for tests.
func ParseLoginLogoDataURLForTest(raw string) (mime string, data []byte, ok bool) {
	return parseLoginLogoDataURL(raw)
}

// LoginBrandLogoGetForTest exposes the public login logo handler.
func LoginBrandLogoGetForTest(w http.ResponseWriter, r *http.Request) {
	LoginBrandLogoHandler(w, r)
}

// ResetAuthTemplateCacheForTest clears parsed auth HTML templates (tests change TemplatesPath).
func ResetAuthTemplateCacheForTest() { resetAuthTemplateCache() }

// LogoutGetForTest exposes the logout GET handler for tests.
func LogoutGetForTest(w http.ResponseWriter, r *http.Request) { LogoutGet(w, r) }

// LogoutPostForTest exposes the logout POST handler for tests.
func LogoutPostForTest(w http.ResponseWriter, r *http.Request) { LogoutPost(w, r) }

// TestSessionTouchRoute is POST /web/session/touch for browser presence.
const TestSessionTouchRoute = sessionTouchRoute

// SessionTouchPostForTest exposes the session touch handler for tests.
func SessionTouchPostForTest(w http.ResponseWriter, r *http.Request) { SessionTouchPost(w, r) }

// BrowserSeenExpiredForTest reports whether browser_seen_at is past configured grace.
func BrowserSeenExpiredForTest(seenValid bool, seenUnix int64) bool {
	var seen sql.NullTime
	if seenValid {
		seen = sql.NullTime{Time: time.Unix(seenUnix, 0).UTC(), Valid: true}
	}
	return browserSeenExpired(seen, time.Now().UTC())
}

// CSRFTokenForRequestForTest exposes the session-bound CSRF token for tests.
func CSRFTokenForRequestForTest(r *http.Request) string { return CSRFTokenForRequest(r) }

// ValidateProductionCSRFSecretForTest exposes production CSRF secret validation.
func ValidateProductionCSRFSecretForTest() error { return ValidateProductionCSRFSecret() }

// LoginPostForTest exposes the login POST handler for tests.
func LoginPostForTest(w http.ResponseWriter, r *http.Request) { LoginPost(w, r) }

// LoginLockedForTest exposes login brute-force lockout state.
func LoginLockedForTest(login string) bool { return loginLocked(login) }

// RecordLoginFailureForTest records a failed login attempt for lockout tests.
func RecordLoginFailureForTest(login string) { recordLoginFailure(login) }

// ClearLoginFailuresForTest clears lockout state for a login key.
func ClearLoginFailuresForTest(login string) { clearLoginFailures(login) }

// ResetLoginLockoutForTest clears all in-memory login lockout state.
func ResetLoginLockoutForTest() { resetLoginLockoutState() }

// ComparePasswordConstantTimeForTest exposes constant-time password compare for tests.
func ComparePasswordConstantTimeForTest(storedHash, plain string) bool {
	return comparePasswordConstantTime(storedHash, plain)
}

// RequireSystemAdminForTest exposes requireSystemAdmin for handler tests.
func RequireSystemAdminForTest(w http.ResponseWriter, r *http.Request, redirectOnDeny bool) bool {
	return requireSystemAdmin(w, r, redirectOnDeny)
}

// RequireModuleManagerOrSystemForTest exposes requireModuleManagerOrSystem for handler tests.
func RequireModuleManagerOrSystemForTest(w http.ResponseWriter, r *http.Request, redirectOnDeny bool) bool {
	return requireModuleManagerOrSystem(w, r, redirectOnDeny)
}

// RequireModelAccessForTest exposes requireModelAccess for handler tests.
func RequireModelAccessForTest(w http.ResponseWriter, r *http.Request, model, perm string) bool {
	return requireModelAccess(w, r, model, perm)
}

// ValidateLoginCSRFForTest exposes pre-session login CSRF validation for tests.
func ValidateLoginCSRFForTest(r *http.Request) bool { return validateLoginCSRF(r) }

// MintSignedUIDCookieForTest builds a signed uid cookie value for tests.
func MintSignedUIDCookieForTest(uid int, ttl time.Duration) string {
	return mintSignedUIDCookieValue(uid, ttl)
}

// ParseSignedUIDCookieForTest parses a signed uid cookie value for tests.
func ParseSignedUIDCookieForTest(value string) (int, bool) {
	return parseSignedUIDCookie(value)
}

// RateLimitedPathForTest reports whether a path is subject to IP rate limiting.
func RateLimitedPathForTest(path string) bool { return rateLimitedPath(path) }

// TestLoginCSRFCookie is the HttpOnly login CSRF cookie name.
const TestLoginCSRFCookie = loginCSRFCookie

// ResolveSessionFromCookieForTest exposes session resolution for integration tests.
func ResolveSessionFromCookieForTest(r *http.Request) (userID int, clearCookie bool) {
	state := resolveSession(r)
	return state.userID, state.clearCookie
}

// TestSessionCookieName is the HttpOnly session cookie name.
const TestSessionCookieName = sessionCookieName

// InsertTestSessionForTest inserts a sys.session row for integration tests.
func InsertTestSessionForTest(sid string, userID int, expiresAt time.Time) error {
	if orm.DB == nil {
		return fmt.Errorf("no database")
	}
	sessionTable := orm.MustQuotedTableName("sys.session")
	_, err := orm.DB.Exec(
		`INSERT INTO `+sessionTable+` (sid, user_id, expires_at) VALUES ($1, $2, $3)`,
		sid, userID, expiresAt,
	)
	return err
}

// CountTestSessionsForUserForTest returns session row count for a user.
func CountTestSessionsForUserForTest(userID int) (int, error) {
	if orm.DB == nil {
		return 0, fmt.Errorf("no database")
	}
	sessionTable := orm.MustQuotedTableName("sys.session")
	var count int
	err := orm.DB.QueryRow(`SELECT COUNT(*) FROM `+sessionTable+` WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

// DeleteTestSessionForTest removes a session row by sid.
func DeleteTestSessionForTest(sid string) error {
	if orm.DB == nil {
		return fmt.Errorf("no database")
	}
	sessionTable := orm.MustQuotedTableName("sys.session")
	_, err := orm.DB.Exec(`DELETE FROM `+sessionTable+` WHERE sid = $1`, sid)
	return err
}

func ResolveExtraScripts(pageScripts, optScripts []string) []string {
	return resolveExtraScripts(pageScripts, optScripts)
}

func PartitionListSectionsForTest(rows []map[string]interface{}, groupField string) []render.ListSection {
	return partitionListSections(rows, groupField)
}

const (
	WorkspaceViewModeFormForTest = workspaceViewModeForm
	WorkspaceViewModeListForTest = workspaceViewModeList
)

func HomeRouteWithMenuForTest(menuID string) string { return homeRouteWithMenu(menuID) }

func PrependViewModeForTest(mode string, modes []string) []string {
	return prependViewMode(mode, modes)
}

func IsNumericRecordIDForTest(recordID string) bool { return isNumericRecordID(recordID) }

func WorkspaceViewModeCandidatesForTest(r *http.Request, actionData map[string]interface{}) []string {
	return workspaceViewModeCandidates(r, actionData)
}

func ActionViewModesForTabsForTest(actionData map[string]interface{}) []string {
	return actionViewModesForTabs(actionData)
}

func ParsePositiveRecordIDForTest(recordIDRaw string) (int, bool) {
	return parsePositiveRecordID(recordIDRaw)
}

type BusHubForTest struct{ hub *busHub }

func NewBusHubForTest() *BusHubForTest {
	return &BusHubForTest{hub: &busHub{clients: make(map[*swcBusClient]struct{})}}
}

type SwcBusClientForTest struct{ client *swcBusClient }

func NewSwcBusClientForTest(uid int, buffer int) *SwcBusClientForTest {
	return &SwcBusClientForTest{client: &swcBusClient{
		uid: uid, send: make(chan []byte, buffer), channels: make(map[string]struct{}),
	}}
}

func (h *BusHubForTest) Register(c *SwcBusClientForTest) { h.hub.register(c.client) }

func (h *BusHubForTest) PublishChannel(channel string, msg []byte) {
	h.hub.publishChannel(channel, msg)
}

func (c *SwcBusClientForTest) SubscribeChannel(channel string) {
	c.client.subMu.Lock()
	c.client.channels[channel] = struct{}{}
	c.client.subMu.Unlock()
}

func (c *SwcBusClientForTest) Recv() <-chan []byte { return c.client.send }

func BuildIframeSwcPayloadForTest(ctx context.Context, actionID int, menuID, iframeURL string) map[string]interface{} {
	p := buildIframeSwcPayload(ctx, actionID, menuID, iframeURL)
	return map[string]interface{}{
		"viewType":  p.ViewType,
		"model":     p.Model,
		"iframeUrl": p.IframeURL,
		"actionId":  p.ActionID,
		"menuId":    p.MenuID,
	}
}

const (
	NavActionURLForTest    = navActionURL
	NavActionWindowForTest = navActionWindow
)

func ResolveNavigationActionForTest(ctx context.Context, actionID int, actionQuery string) (kind int, url string, err error) {
	nav, err := resolveNavigationAction(ctx, actionID, actionQuery)
	if err != nil {
		return 0, "", err
	}
	switch nav.kind {
	case navActionURL:
		return int(navActionURL), nav.url, nil
	case navActionWindow:
		return int(navActionWindow), "", nil
	default:
		return 0, "", err
	}
}

// ShareTokenParsed is the decoded portal share token payload.
type ShareTokenParsed struct {
	IssuerUID int
	Model     string
	ResID     int64
}

// MintShareTokenForTest signs a share link for tests.
func MintShareTokenForTest(issuerUID int, model string, resID int64) (string, error) {
	return MintShareToken(issuerUID, model, resID)
}

// MintShareTokenForTestWithExpiry signs a share link with a fixed expiry unix time.
func MintShareTokenForTestWithExpiry(issuerUID int, model string, resID int64, expUnix int64) (string, error) {
	return mintShareTokenAt(issuerUID, model, resID, expUnix)
}

// ParseShareTokenForTest validates and decodes a share token.
func ParseShareTokenForTest(token string) (ShareTokenParsed, error) {
	parsed, err := parseShareToken(token)
	if err != nil {
		return ShareTokenParsed{}, err
	}
	return ShareTokenParsed{
		IssuerUID: parsed.IssuerUID,
		Model:     parsed.Model,
		ResID:     parsed.ResID,
	}, nil
}

func FieldACLMatrixRowNameForTest(model, field string, groupID int) string {
	return fieldACLMatrixRowName(model, field, groupID)
}

func ModelACLMatrixRowNameForTest(model string, groupID int) string {
	return modelACLMatrixRowName(model, groupID)
}

func MatrixGroupDisplayLenForTest(groupCount int, showAll bool) (displayLen int, truncated bool) {
	groups := make([]aclGroupCol, groupCount)
	for i := range groups {
		groups[i] = aclGroupCol{ID: i + 1, Label: "g"}
	}
	out, truncated := matrixGroupsForDisplay(groups, showAll)
	return len(out), truncated
}

func FilterFieldNamesForTest(fields []string, filter string) []string {
	return filterFieldNames(fields, filter)
}

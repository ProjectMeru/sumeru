package web

import (
	"context"
	"net/http"
	"time"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/report"
)

// Type aliases for external tests in core/server/web (web_test package).
type (
	ShellPageOpts      = shellPageOpts
	SetupInitRequest   = setupInitRequest
	AppsBrowseState    = appsBrowseState
	ModuleRow          = moduleRow
	PinnedAppsRequest  = pinnedAppsRequest
	WorkspaceRequest   = workspaceRequest
	AppsModule         = appsModule
	AppsModuleGroup    = appsModuleGroup
	SettingsHubSection = settingsHubSection
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
	TestPinnedAppsRoute        = pinnedAppsRoute
	TestChatterPostRoute       = chatterPostRoute
	TestCompanySwitchRoute     = companySwitchRoute
	TestModuleActionRoute      = moduleActionRoute
	TestAPIRPCRoute            = apiRPCRoute
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
	TestWorkspaceStylesheetURL   = workspaceStylesheetURL
	TestPagesStylesheetURL       = pagesStylesheetURL
	TestSettingsHubStylesheetURL = settingsHubStylesheetURL
	TestMaxRPCBodyBytes          int64 = maxRPCBodyBytes
	TestMaxChatterBodyRunes      = maxChatterBodyRunes
	TestSetupRateLimitWindow     = setupRateLimitWindow
	TestSetupRateLimitMax        = setupRateLimitMax
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
	return pruneSetupAttempts(attempts, now)
}

func AllowSetupRateLimit(w http.ResponseWriter, requestIP string) bool {
	return allowSetupRateLimit(w, requestIP)
}

func ValidateSetupToken(w http.ResponseWriter, r *http.Request, tokenFromBody string) bool {
	return validateSetupToken(w, r, tokenFromBody)
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

func SettingsHubSectionFromSidebar(sidebarSection render.SidebarMenu) (SettingsHubSection, bool) {
	return settingsHubSectionFromSidebar(sidebarSection)
}

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

func ParseChatterPostForm(r *http.Request) chatterPostForm { return parseChatterPostForm(r) }

func ChatterBodyTooLong(body string) bool { return chatterBodyTooLong(body) }

func ParseChatterRecordID(recordIDRaw string) (int64, error) {
	return parseChatterRecordID(recordIDRaw)
}

func CoerceCSVValue(raw string) interface{} { return coerceCSVValue(raw) }

func ImportableRowValues(header []string, record []string, allowedFields map[string]struct{}) map[string]interface{} {
	return importableRowValues(header, record, allowedFields)
}

func IsImportableColumn(column string, allowedFields map[string]struct{}) bool {
	return isImportableColumn(column, allowedFields)
}

func ImportCSVFlashMessage(createdCount int) string { return importCSVFlashMessage(createdCount) }

func ParseCompanySwitchForm(r *http.Request) companySwitchForm { return parseCompanySwitchForm(r) }

func LoginURLWithReturn(returnTo string) string { return loginURLWithReturn(returnTo) }

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
	return appsDetailURL(browse, editing)
}

func FindAppsModule(modules []AppsModule, moduleName string) (AppsModule, bool) {
	return findAppsModule(modules, moduleName)
}

func APIKeyTargetUserID(r *http.Request) int { return apiKeyTargetUserID(r) }

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

// SetTestSessionUserIDForTest overrides SessionUserID for handler tests.
func SetTestSessionUserIDForTest(userID int) { testSessionUserIDOverride = userID }

// ResetTestSessionUserIDForTest clears the session override.
func ResetTestSessionUserIDForTest() { testSessionUserIDOverride = 0 }

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

func PrependViewModeForTest(mode string, modes []string) []string { return prependViewMode(mode, modes) }

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
	return &SwcBusClientForTest{client: &swcBusClient{uid: uid, send: make(chan []byte, buffer)}}
}

func (h *BusHubForTest) Register(c *SwcBusClientForTest) { h.hub.register(c.client) }

func (h *BusHubForTest) Broadcast(actor int, msg []byte) { h.hub.broadcast(actor, msg) }

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

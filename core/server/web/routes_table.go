// Package web implements HTTP handlers for the signed-in UI (/web/*), setup, and JSON APIs.
package web

import (
	"net/http"

	"sumeru/core/server/router"

	_ "sumeru/core/report" // register bulk import object actions
)

func init() {
	wireRouterAuth()
}

func wireRouterAuth() {
	router.RequireSession = func(w http.ResponseWriter, r *http.Request) bool {
		return requireLogin(w, r)
	}
	router.ResolveUID = func(r *http.Request) int {
		return AuthenticatedUserID(r)
	}
}

// RegisterAppRoutes registers HTTP handlers for /web and related paths after DB init.
func RegisterAppRoutes(mux *http.ServeMux) {
	serveMux := serveMuxOrDefault(mux)

	registerAuthRoutes()
	registerWorkspaceRoutes()
	registerSwcRoutes()
	registerAppsRoutes()
	registerRecordRoutes()
	registerActionRoutes()
	registerSettingsRoutes()
	registerAPIRoutes()

	router.Apply(serveMux)
	registerAppAliases(serveMux)
}

func registerAuthRoutes() {
	registerPublic(http.MethodGet, loginRoute, LoginGet)
	registerPublic(http.MethodPost, loginRoute, LoginPost)
	registerPublic(http.MethodGet, logoutRoute, LogoutGet)
}

func registerWorkspaceRoutes() {
	registerSession(http.MethodGet, homeRoute, HomeDashboardHandler)
	registerSession(http.MethodPost, pinnedAppsRoute, PinnedAppsSaveHandler)
	registerSession(http.MethodGet, workspaceRoute, WebHandler)
	registerSession(http.MethodPost, companySwitchRoute, SwitchCompanyPost)
	registerSession(http.MethodPost, chatterPostRoute, ChatterPostHandler)
	registerSession(http.MethodPost, importCSVRoute, ImportCSVHandler)
}

func registerAppsRoutes() {
	registerSession(http.MethodGet, appsRoute, AppsHandler)
	registerSession(http.MethodPost, moduleActionRoute, ModuleActionHandler)
}

func registerRecordRoutes() {
	registerReportRoutes()
	registerSession(http.MethodGet, exportCSVRoute, ExportCSVHandler)
	registerSession(http.MethodGet, exportPDFRoute, ExportPDFHandler)
	registerSession(http.MethodGet, exportXLSXRoute, ExportXLSXHandler)
	registerSession(http.MethodGet, templatePDFRoute, ExportTemplatePDFHandler)
	registerSession(http.MethodGet, bulkTemplateRoute, BulkTemplateHandler)
	registerSession(http.MethodPost, bulkUploadRoute, BulkUploadHandler)
	registerSession(http.MethodPost, bulkConfirmRoute, BulkConfirmHandler)
	registerSession(http.MethodPost, bulkCancelRoute, BulkCancelHandler)
}

func registerActionRoutes() {
	registerSession(http.MethodPost, resetPasswordRoute, ActionResetPassword)
	registerSession(http.MethodPost, createAPIKeyRoute, ActionCreateAPIKey)
}

func registerSettingsRoutes() {
	registerSession(http.MethodGet, settingsRoute, SettingsHubHandler)
	registerSession(http.MethodGet, appLogsRoute, AppLogsHandler)
	registerSession(http.MethodGet, metricsRoute, MetricsHandler)
}

func registerAPIRoutes() {
	registerPublic(http.MethodGet, apiHealthRoute, APIHealthHandler)
	registerPublic(http.MethodGet, apiReadyRoute, APIReadyHandler)
	registerPublic(http.MethodPost, apiRPCRoute, RPCJSONHandler)
}

func registerPublic(method, path string, handler http.HandlerFunc) {
	router.Register(method, path, router.AuthNone, handler)
}

func registerSession(method, path string, handler http.HandlerFunc) {
	router.Register(method, path, router.AuthSession, handler)
}

func serveMuxOrDefault(mux *http.ServeMux) *http.ServeMux {
	if mux != nil {
		return mux
	}
	return http.DefaultServeMux
}

func registerAppAliases(mux *http.ServeMux) {
	mux.HandleFunc(appsRoute+"/", redirectFoundTo(appsRoute))
	mux.HandleFunc(settingsRoute+"/", redirectFoundTo(settingsRoute))
	mux.HandleFunc(rootRoute, rootRedirectHome)
}

func redirectFoundTo(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
	}
}

func rootRedirectHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != rootRoute {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, homeRoute, http.StatusFound)
}

// RegisterSetupRoutes registers first-run setup handlers (DB not ready).
func RegisterSetupRoutes(mux *http.ServeMux) {
	serveMux := serveMuxOrDefault(mux)
	serveMux.HandleFunc(setupRoute, SetupPageHandler)
	serveMux.HandleFunc(setupInitRoute, SetupInitHandler)
	serveMux.HandleFunc(rootRoute, rootRedirectSetup)
}

func rootRedirectSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != rootRoute {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, setupRoute, http.StatusFound)
}

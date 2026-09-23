package assets

// DefaultStylesheetURLs is the ordered core CSS stack for authenticated shell pages.
// sumeru-ai.css is inserted after layout slices when sumeru_ai is installed (see static.go).
// Login and app-logs sheets are page-specific — use LoginStylesheetURLs / AppLogsStylesheetURLs.
func DefaultStylesheetURLs() []string {
	return []string{
		"/static/css/sumeru-theme.css",
		"/static/css/sumeru-base.css",
		"/static/css/sumeru-shell.css",
		"/static/css/sumeru-messages.css",
		"/static/css/sumeru-views.css",
		"/static/css/sumeru-compat.css",
		"/static/css/sumeru-apps.css",
	}
}

// LoginStylesheetURLs extends the core stack for login and setup wizard pages.
func LoginStylesheetURLs() []string {
	urls := DefaultStylesheetURLs()
	return append(urls, "/static/css/sumeru-login.css")
}

// AppLogsStylesheetURLs extends workspace styles with pages layout for app logs.
func AppLogsStylesheetURLs() []string {
	return []string{"/static/css/sumeru-workspace.css", "/static/css/sumeru-pages.css"}
}

// PortalStylesheetURLs is the minimal CSS stack for /portal pages.
func PortalStylesheetURLs() []string {
	return []string{
		"/static/css/sumeru-theme.css",
		"/static/css/sumeru-base.css",
		"/static/css/portal.css",
	}
}

// AIStylesheetURL is the optional AI assistant stylesheet path.
func AIStylesheetURL() string {
	return "/static/css/sumeru-ai.css"
}

package render

// ExtraStylesheetURLs are linked after view stylesheets (e.g. /static/brand.css from deployment).
var ExtraStylesheetURLs []string

// ExtraScriptURLs are deferred script tags injected before the SWC bundle.
var ExtraScriptURLs []string

// SetExtraStylesheetURLs replaces the list of extra stylesheet URLs (absolute paths on site).
func SetExtraStylesheetURLs(urls []string) {
	ExtraStylesheetURLs = append([]string(nil), urls...)
}

// SetExtraScriptURLs replaces deferred script URLs for the shell layout.
func SetExtraScriptURLs(urls []string) {
	ExtraScriptURLs = append([]string(nil), urls...)
}

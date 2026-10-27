package render

import (
	"regexp"
	"strings"
)

const shellWebPrefix = "/web/"

var moduleNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ReservedWebSegments are /web/{segment} paths owned by the kernel, not application modules.
var ReservedWebSegments = map[string]struct{}{
	"login": {}, "logout": {}, "home": {}, "apps": {}, "settings": {},
	"import": {}, "export": {}, "bulk": {}, "swc": {}, "debug": {},
	"company": {}, "module": {}, "action": {}, "session": {}, "report": {},
	"kanban": {}, "setup": {}, "portal": {}, "api": {}, "user": {}, "content": {},
	"auth": {}, "apikey": {},
}

// AppsWebURL is the installed-apps administration catalog.
func AppsWebURL() string {
	return "/web/apps"
}

// ModuleWebURL returns /web/{moduleName} when the name is a valid, non-reserved module segment.
func ModuleWebURL(moduleName string) string {
	moduleName = strings.TrimSpace(moduleName)
	if !ValidModuleWebSegment(moduleName) {
		return ""
	}
	return shellWebPrefix + moduleName
}

// ValidModuleWebSegment reports whether segment is safe to use as /web/{segment}.
func ValidModuleWebSegment(segment string) bool {
	segment = strings.TrimSpace(segment)
	if segment == "" || !moduleNameRE.MatchString(segment) {
		return false
	}
	_, reserved := ReservedWebSegments[strings.ToLower(segment)]
	return !reserved
}

// ParseDebugQuery interprets the debug query param: true, or true/{moduleName}.
func ParseDebugQuery(debugParam string) (enabled bool, moduleName string) {
	debugParam = strings.TrimSpace(debugParam)
	if debugParam == "" {
		return false, ""
	}
	if debugParam == "true" {
		return true, ""
	}
	const prefix = "true/"
	if strings.HasPrefix(debugParam, prefix) {
		mod := strings.TrimSpace(debugParam[len(prefix):])
		if ValidModuleWebSegment(mod) {
			return true, mod
		}
		return true, ""
	}
	return false, ""
}

// DebugQueryValue builds the debug query value for a module (debug=true/crm).
func DebugQueryValue(moduleName string) string {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return "true"
	}
	return "true/" + moduleName
}

// WebDebugQueryURL is GET /web?debug=true or /web?debug=true/{module}.
func WebDebugQueryURL(moduleName string) string {
	q := DebugQueryValue(moduleName)
	return WorkspaceRoute + "?debug=" + q
}

// ModuleWebURLWithDebugAlias returns /web/{module}?debug=true (alias for module entry with debug).
func ModuleWebURLWithDebugAlias(moduleName string) string {
	base := ModuleWebURL(moduleName)
	if base == "" {
		return ""
	}
	return base + "?debug=true"
}

package render

import (
	"encoding/json"
	"html/template"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/module"
	"sumeru/core/orm"
)

// BuildSwcAddonEntriesJSON returns JSON array of SWC entry script URLs for installed addons.
func BuildSwcAddonEntriesJSON() template.JS {
	urls := collectInstalledSwcEntryURLs()
	if len(urls) == 0 {
		return template.JS("[]")
	}
	raw, err := json.Marshal(urls)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(raw)
}

func collectInstalledSwcEntryURLs() []string {
	names := make([]string, 0, len(module.LoadedAddons))
	for name := range module.LoadedAddons {
		names = append(names, name)
	}
	sort.Strings(names)

	var urls []string
	seen := map[string]struct{}{}
	for _, moduleName := range names {
		addon := module.LoadedAddons[moduleName]
		if addon == nil || !addonModuleInstalled(moduleName) {
			continue
		}
		entry := strings.TrimSpace(addon.Manifest.SwcEntry)
		if entry == "" {
			continue
		}
		publicURL := swcEntryPublicURL(moduleName, entry)
		if publicURL == "" {
			continue
		}
		if _, ok := seen[publicURL]; ok {
			continue
		}
		seen[publicURL] = struct{}{}
		urls = append(urls, publicURL)
	}
	return urls
}

func swcEntryPublicURL(moduleName, entry string) string {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return ""
	}
	if strings.HasPrefix(entry, "/") {
		return entry
	}
	entry = filepath.ToSlash(entry)
	if entry == "" || strings.Contains(entry, "..") {
		return ""
	}
	return "/static/addon-asset/" + moduleName + "/" + entry
}

func addonModuleInstalled(moduleName string) bool {
	if orm.DB == nil {
		return false
	}
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return false
	}
	table := orm.MustQuotedTableName("sys.module")
	var state string
	err := orm.DB.QueryRow(
		`SELECT state FROM `+table+` WHERE name = $1 AND active = true`,
		moduleName,
	).Scan(&state)
	return err == nil && strings.TrimSpace(state) == "installed"
}

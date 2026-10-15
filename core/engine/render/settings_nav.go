package render

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// MenuGeneralSettingsXMLID is the settings hub sidebar entry (URL action → SettingsHubPath).
const MenuGeneralSettingsXMLID = "base.menu_general_settings"

const (
	SettingsHubPath      = "/web/settings"
	SettingsAccountPath  = "/web/settings/account"
	SettingsFieldACLPath = "/web/settings/field-acl"
	SettingsModelACLPath = "/web/settings/model-acl"
)

// Menu XML ids for settings security matrix pages (active sidebar highlight).
const (
	MenuFieldAccessMatrixXMLID = "base.menu_sys_field_access_matrix"
	MenuModelAccessMatrixXMLID = "base.menu_sys_access_matrix"
)

func settingsNavExcludedSection(title string) bool {
	return strings.EqualFold(strings.TrimSpace(title), "Personal")
}

func settingsNavExcludedLink(name, href string) bool {
	if strings.EqualFold(strings.TrimSpace(name), "Account security") {
		return true
	}
	return strings.Contains(strings.TrimSpace(href), SettingsAccountPath)
}

func menuIDForXMLID(ctx context.Context, xmlID string) string {
	xmlID = strings.TrimSpace(xmlID)
	if xmlID == "" {
		return ""
	}
	id, _, err := orm.ResolveXmlId(ctx, xmlID)
	if err != nil || 0 >= id {
		return ""
	}
	return strconv.Itoa(id)
}

func isSettingsHubNavMenu(ctx context.Context, menuID string) bool {
	menuID = strings.TrimSpace(menuID)
	if menuID == "" {
		return false
	}
	return menuID == menuIDForXMLID(ctx, MenuGeneralSettingsXMLID)
}

// MenuIDIsGeneralSettings reports whether menuID is base.menu_general_settings.
func MenuIDIsGeneralSettings(ctx context.Context, menuID string) bool {
	return isSettingsHubNavMenu(ctx, menuID)
}

// MenuIDForXMLID resolves a menu xml id to the numeric menu id string used in the shell.
func MenuIDForXMLID(ctx context.Context, xmlID string) string {
	return menuIDForXMLID(ctx, xmlID)
}

func applySettingsNavHrefOverrides(ctx context.Context, menus []parser.MenuItem) {
	type override struct {
		xmlID string
		href  string
	}
	for _, o := range []override{
		{MenuGeneralSettingsXMLID, SettingsHubPath},
		{MenuFieldAccessMatrixXMLID, SettingsFieldACLPath},
		{MenuModelAccessMatrixXMLID, SettingsModelACLPath},
	} {
		menuID := menuIDForXMLID(ctx, o.xmlID)
		if menuID == "" {
			continue
		}
		for i := range menus {
			if menus[i].ID == menuID {
				menus[i].Action = o.href
				break
			}
		}
	}
}

// ApplySettingsNavHrefOverridesForTest applies in-memory href overrides (tests without DB xml ids).
func ApplySettingsNavHrefOverridesForTest(menus []parser.MenuItem, generalSettingsMenuID string) {
	generalSettingsMenuID = strings.TrimSpace(generalSettingsMenuID)
	if generalSettingsMenuID == "" {
		return
	}
	for i := range menus {
		if menus[i].ID == generalSettingsMenuID {
			menus[i].Action = SettingsHubPath
			return
		}
	}
}

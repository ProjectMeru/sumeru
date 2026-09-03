package render

import (
	"sumeru/core/engine/parser"
)

func BuildSidebarMenus(allMenus []parser.MenuItem, rootMenuID string, menuAllowed func(parser.MenuItem) bool) []SidebarMenu {
	return buildSidebarMenus(allMenus, rootMenuID, menuAllowed)
}

func ResolveActiveModuleID(allMenus []parser.MenuItem, activeMenuID string) string {
	return resolveActiveModuleID(allMenus, activeMenuID)
}

type MenuCrumbForTest = menuCrumb

func SplitModuleMenuChainForTest(chain []menuCrumb) (module menuCrumb, menus []menuCrumb, ok bool) {
	return splitModuleMenuChain(chain)
}

func WorkspaceViewBreadcrumbLabelForTest(viewType string) string {
	return workspaceViewBreadcrumbLabel(viewType)
}

func WorkspaceRecordBreadcrumbLabelForTest(in BreadcrumbInput) string {
	return workspaceRecordBreadcrumbLabel(in)
}

func ViewModeFilterSetForTest(modes []string) map[string]struct{} {
	return viewModeFilterSet(modes)
}

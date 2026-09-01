package render

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/engine/parser"
)

// RenderSWCWorkspace builds the shell HTML page with an empty SWC mount point.
func RenderSWCWorkspace(ctx context.Context, view *parser.View, activeMenuID, templatesDir string, recData *ViewRecordData, selectedMode string) string {
	if recData == nil {
		recData = &ViewRecordData{}
	}
	content := `<div id="swc-workspace" class="sum-swc-mount" aria-live="polite"></div>`

	topMenus, sidebarMenus, activeModuleID, moduleName := LoadShellMenus(ctx, activeMenuID)
	viewBC := HumanViewBreadcrumb(view.Model, selectedMode)

	actCtxModel := ""
	var actCtxID int64
	if strings.EqualFold(selectedMode, ViewModeForm) && recData.RecordID > 0 {
		actCtxModel = strings.TrimSpace(view.Model)
		actCtxID = int64(recData.RecordID)
	}

	pageData := PageData{
		Title:                   fmt.Sprintf("%s · %s", view.Model, selectedMode),
		ViewBreadcrumb:          viewBC,
		ModuleName:              moduleName,
		Content:                 template.HTML(content),
		TopMenus:                topMenus,
		SidebarMenus:            sidebarMenus,
		ActiveModuleID:          activeModuleID,
		ActiveMenuID:            activeMenuID,
		ViewStylesheetURLs:      []string{"/static/css/sumeru-workspace.css"},
		ExtraStylesheetURLs:     ExtraStylesheetURLs,
		ExtraScriptURLs:         ExtraScriptURLs,
		ViewTabs:                recData.ViewTabs,
		ActivityContextModel:    actCtxModel,
		ActivityContextRecordID: actCtxID,
		SettingsNavActive:       IsMenuUnderSettingsRoot(ctx, activeMenuID),
		BreadcrumbItems: BuildWorkspaceBreadcrumbs(ctx, BreadcrumbInput{
			ActiveMenuID:  activeMenuID,
			ViewType:      selectedMode,
			Title:         viewBC,
			FormBaseQuery: recData.FormBaseQuery,
			Record:        recData.Record,
			RecordID:      recData.RecordID,
		}),
		CSRFToken:               recData.CSRFToken,
		SWCEnabled:              true,
	}
	inlineFlashes, toastFlashes := splitFlashMessages(recData.FlashMessages)
	pageData.FlashMessages = inlineFlashes
	pageData.ToastMessages = toastFlashes
	if len(toastFlashes) > 0 {
		if b, err := json.Marshal(toastFlashes); err == nil {
			pageData.ToastMessagesJSON = template.JS(b)
		}
	}
	ws := &SWCBootstrapWorkspace{
		ActionID:   recData.ActionID,
		MenuID:     activeMenuID,
		ViewType:   selectedMode,
		RecordID:   recData.RecordID,
		FormEdit:   recData.FormEditing,
		ListSearch: recData.ListSearchQuery,
	}
	EnrichShellPageData(ctx, &pageData)
	pageData.SWCBootstrapJSON = BuildSWCBootstrapJSON(ctx, pageData, ws)
	if !SidebarHasMenus(sidebarMenus) {
		pageData.SuppressSidebar = true
	}

	applog.DebugMsg(ctx, "render", "swc",
		fmt.Sprintf("Rendering SWC workspace for model %s", view.Model),
		map[string]interface{}{"active_menu": activeMenuID, "view_type": selectedMode})

	out, err := RenderPage(ctx, templatesDir, pageData)
	if err != nil {
		applog.WarnMsg(ctx, "render", "swc", "Error rendering SWC page", err,
			map[string]interface{}{"model": view.Model})
		return content
	}
	return out
}

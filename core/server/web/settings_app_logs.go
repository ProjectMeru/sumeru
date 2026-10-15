package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

// AppLogsHandler serves Event Log inside the main shell.
func AppLogsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireMenuAccess(w, r, appLogsMenuXMLID) {
		return
	}
	uid := SessionUserID(r)
	if err := orm.CheckModelAccess(r.Context(), uid, appLogModel, "read"); err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	events, err := loadAppLogEvents(ctx)
	if err != nil {
		WebLogEvent(ctx, WebLogInput{
			Route: appLogsRoute, Message: "Failed to load app logs",
			Operation: "load", Status: "failure", Err: err,
		})
		http.Error(w, "Failed to load app logs", http.StatusInternalServerError)
		return
	}

	menuID, menuIDStr := resolveMenuID(ctx, appLogsMenuXMLID)
	renderAppLogsPage(w, r, events, menuID, menuIDStr)
}

type appLogEvent struct {
	CreatedAt string
	Module    string
	Action    string
	Detail    string
}

func renderAppLogsPage(w http.ResponseWriter, r *http.Request, events []appLogEvent, menuID int, menuIDStr string) {
	ctx := r.Context()
	renderShellPage(w, r, shellPageOpts{
		Route:         appLogsRoute,
		InnerTemplate: appLogsInnerTemplate,
		InnerData:     events,
		MenuIDStr:     menuIDStr,
		Page:          buildAppLogsPageData(ctx, menuID),
	})
}

func buildAppLogsPageData(ctx context.Context, menuID int) render.PageData {
	page := render.PageData{
		Title:              appLogsPageTitle,
		ViewBreadcrumb:     appLogsBreadcrumb,
		SettingsNavActive:  true,
		ViewStylesheetURLs: appLogsViewStylesheets(),
	}
	if menuID > 0 {
		page.BreadcrumbItems = render.BuildAppLogsBreadcrumbs(ctx, menuID)
	}
	return page
}

func appLogsViewStylesheets() []string {
	return assets.AppLogsStylesheetURLs()
}

func resolveMenuID(ctx context.Context, menuXMLID string) (menuID int, menuIDStr string) {
	menuID, ok := resolvedMenuIDFromXML(ctx, menuXMLID)
	if !ok {
		return 0, ""
	}
	return menuID, strconv.Itoa(menuID)
}

func loadAppLogEvents(ctx context.Context) ([]appLogEvent, error) {
	rows, err := orm.SearchPage(ctx, appLogModel, nil, maxAppLogEvents, 0, "create_date desc, id desc")
	if err != nil {
		return nil, err
	}
	events := make([]appLogEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, appLogEvent{
			CreatedAt: formatAppLogTimestamp(row["create_date"]),
			Module:    strings.TrimSpace(orm.AsString(row["module_name"])),
			Action:    strings.TrimSpace(orm.AsString(row["action"])),
			Detail:    strings.TrimSpace(orm.AsString(row["detail"])),
		})
	}
	return events, nil
}

func formatAppLogTimestamp(raw interface{}) string {
	s := strings.TrimSpace(orm.AsString(raw))
	if len(s) >= 19 {
		return s[:10] + " " + s[11:19]
	}
	return s
}

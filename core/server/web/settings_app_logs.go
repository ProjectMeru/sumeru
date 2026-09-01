package web

import (
	"context"
	"net/http"
	"strconv"

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
	logTable := orm.MustQuotedTableName(appLogModel)
	query := `
		SELECT
			to_char(create_date, 'YYYY-MM-DD HH24:MI:SS') AS created_at,
			COALESCE(NULLIF(TRIM(module_name), ''), '') AS module,
			COALESCE(NULLIF(TRIM(action), ''), '') AS action,
			COALESCE(NULLIF(TRIM(detail), ''), '') AS detail
		FROM ` + logTable + `
		ORDER BY create_date DESC, id DESC
		LIMIT ` + strconv.Itoa(maxAppLogEvents)

	rows, err := orm.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]appLogEvent, 0, maxAppLogEvents)
	for rows.Next() {
		var event appLogEvent
		if err := rows.Scan(&event.CreatedAt, &event.Module, &event.Action, &event.Detail); err != nil {
			WebLogEvent(ctx, WebLogInput{
				Route: appLogsRoute, Message: "App log row scan failed",
				Operation: "load", Status: "partial", Err: err,
			})
			continue
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

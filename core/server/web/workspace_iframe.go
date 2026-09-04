package web

import (
	"context"
	"net/http"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

func renderURLActionWorkspace(w http.ResponseWriter, r *http.Request, actionID int, menuQuery, iframeURL string) {
	ctx := r.Context()
	menuID := CanonicalMenuID(ctx, menuQuery, actionID)
	title := urlActionTitle(ctx, actionID)
	syntheticView := &parser.View{
		Model: "sys.action.url",
		Type:  render.ViewModeIframe,
		Title: title,
	}
	recData := &render.ViewRecordData{
		ActionID: actionID,
		ResModel: "sys.action.url",
	}
	html := render.RenderSWCWorkspace(ctx, render.SWCPageInput{
		View:         syntheticView,
		ActiveMenuID: menuID,
		TemplatesDir: config.AppConfig.TemplatesPath,
		RecData:      recData,
		SelectedMode: render.ViewModeIframe,
	})
	writeHTML(w, ctx, r.URL.Path, html)
}

func buildIframeSwcPayload(ctx context.Context, actionID int, menuID, iframeURL string) swcmeta.WorkspacePayload {
	return swcmeta.BuildIframeWorkspacePayload(actionID, menuID, iframeURL, urlActionTitle(ctx, actionID))
}

func urlActionTitle(ctx context.Context, actionID int) string {
	if actionID <= 0 {
		return "Report"
	}
	row, err := orm.SearchOne(ctx, sysActionURLModel, map[string]interface{}{"id": actionID})
	if err != nil {
		return "Report"
	}
	if name := strings.TrimSpace(orm.AsString(row["name"])); name != "" {
		return name
	}
	return "Report"
}

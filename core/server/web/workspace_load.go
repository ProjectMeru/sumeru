package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

type workspaceRequest struct {
	actionID    int
	menuID      string
	viewType    string
	recordID    string
	formEdit    bool
	listSearch  string
	model       string
	listFilter  string
	listDomain  string
	listSort    string
	listOffset  int
	listGroupBy string
}

type workspaceLoadInput struct {
	ViewRecord *render.ViewRecordData
	Resolved   *resolvedWorkspaceView
	ActionData map[string]interface{}
	Req        workspaceRequest
}

type searchWorkspaceRowsInput struct {
	workspaceLoadInput
	View        *parser.View
	SearchQuery string
	RowLimit    int
}


func buildViewRecordData(ctx context.Context, w http.ResponseWriter, r *http.Request, req workspaceRequest, resolved *resolvedWorkspaceView, actionData map[string]interface{}) (*render.ViewRecordData, error) {
	viewRecord := &render.ViewRecordData{
		ActionID:    req.actionID,
		CSRFToken:   CSRFTokenForRequest(r),
		ResModel:    resolved.targetModel,
		FormEditing: req.formEdit,
		FormBaseQuery: render.WorkspaceQueryString(render.WorkspaceQuery{
			ActionID: req.actionID,
			MenuID:   req.menuID,
			ViewType: workspaceViewModeForm,
			RecordID: req.recordID,
			Model:    req.model,
		}),
		ViewTabs: render.WorkspaceViewTabs(ctx, render.WorkspaceTabsInput{
			ResModel:     resolved.targetModel,
			ActionID:     req.actionID,
			MenuID:       req.menuID,
			SelectedMode: resolved.selectedMode,
			RecordID:     req.recordID,
			ViewModes:    actionViewModesForTabs(actionData),
		}),
	}
	appendPageFlashesToViewRecord(r, w, viewRecord)
	appendQueryFlashesToViewRecord(r, viewRecord)

	if recordID, ok := parsePositiveRecordID(req.recordID); ok {
		viewRecord.RecordID = recordID
	}

	if err := loadViewModeData(ctx, workspaceLoadInput{
		ViewRecord: viewRecord,
		Resolved:   resolved,
		ActionData: actionData,
		Req:        req,
	}); err != nil {
		return nil, err
	}
	return viewRecord, nil
}

func appendPageFlashesToViewRecord(r *http.Request, w http.ResponseWriter, viewRecord *render.ViewRecordData) {
	for _, flash := range ConsumePageFlashes(r, w) {
		viewRecord.FlashMessages = append(viewRecord.FlashMessages, render.FlashMessage{
			Kind:    flash.Kind,
			Title:   flash.Title,
			Body:    flash.Body,
			Details: flash.Details,
		})
	}
}

func parsePositiveRecordID(recordIDRaw string) (int, bool) {
	recordID, err := strconv.Atoi(strings.TrimSpace(recordIDRaw))
	return recordID, err == nil && recordID > 0
}


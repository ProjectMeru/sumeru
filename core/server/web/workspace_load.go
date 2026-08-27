package web

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
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
	listSort    string
	listOffset  int
	listGroupBy string
}

func parseWorkspaceRequest(r *http.Request, actionID int) workspaceRequest {
	query := r.URL.Query()
	offset, _ := strconv.Atoi(strings.TrimSpace(query.Get(workspaceOffsetParam)))
	if offset < 0 {
		offset = 0
	}
	return workspaceRequest{
		actionID:    actionID,
		menuID:      query.Get(workspaceMenuIDParam),
		viewType:    strings.TrimSpace(query.Get(workspaceViewTypeParam)),
		recordID:    strings.TrimSpace(query.Get(workspaceRecordIDParam)),
		formEdit:    strings.TrimSpace(query.Get(workspaceEditParam)) == workspaceEditEnabledValue,
		listSearch:  listSearchQuery(r),
		model:       strings.TrimSpace(query.Get(workspaceModelParam)),
		listFilter:  strings.TrimSpace(query.Get(workspaceFilterParam)),
		listSort:    strings.TrimSpace(query.Get(workspaceSortParam)),
		listOffset:  offset,
		listGroupBy: strings.TrimSpace(query.Get(workspaceGroupByParam)),
	}
}

// resolveWorkspaceRedirect returns a redirect URL when menu_id has no window action.
func resolveWorkspaceRedirect(ctx context.Context, menuID string, resolvedActionID int) (redirectURL string, shouldRedirect bool) {
	if resolvedActionID != 0 {
		return "", false
	}
	switch {
	case menuIDPointsToAppLogs(ctx, menuID):
		return appLogsRoute, true
	case render.IsMenuUnderSettingsRoot(ctx, menuID):
		return settingsRoute, true
	case isHomeMenuTree(ctx, menuID):
		return homeRouteWithMenu(menuID), true
	default:
		return appsRoute, true
	}
}

func homeRouteWithMenu(menuID string) string {
	if menuID = strings.TrimSpace(menuID); menuID == "" {
		return homeRoute
	}
	return homeRoute + "?" + workspaceMenuIDParam + "=" + menuID
}

type resolvedWorkspaceView struct {
	viewData     map[string]interface{}
	selectedMode string
	view         *parser.View
	targetModel  string
}

func resolveWorkspaceView(ctx context.Context, r *http.Request, actionData map[string]interface{}) (*resolvedWorkspaceView, error) {
	targetModel := actionWindowTargetModel(actionData)
	if targetModel == "" {
		return nil, fmt.Errorf("action has no target model")
	}

	viewModes := workspaceViewModeCandidates(r, actionData)
	viewData, selectedMode, err := findFirstWorkspaceView(ctx, targetModel, viewModes, actionViewIDFromContext(actionData))
	if err != nil {
		return nil, err
	}

	arch := strings.TrimSpace(orm.AsString(viewData["arch"]))
	if arch == "" {
		return nil, fmt.Errorf("view arch is empty")
	}
	parsedView, err := parser.ParseViewFromArch(arch)
	if err != nil {
		return nil, fmt.Errorf("parse view arch: %w", err)
	}

	return &resolvedWorkspaceView{
		viewData:     viewData,
		selectedMode: selectedMode,
		view:         parsedView,
		targetModel:  targetModel,
	}, nil
}

func resolveWorkspaceByModel(ctx context.Context, r *http.Request, model string) (*resolvedWorkspaceView, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("action required")
	}
	uid := orm.SecurityUID(ctx)
	if err := orm.CheckModelAccess(ctx, uid, model, "read"); err != nil {
		return nil, err
	}
	viewType := normalizeViewMode(r.URL.Query().Get(workspaceViewTypeParam))
	if viewType == "" {
		if recordID := strings.TrimSpace(r.URL.Query().Get(workspaceRecordIDParam)); isNumericRecordID(recordID) {
			viewType = workspaceViewModeForm
		} else {
			viewType = workspaceViewModeList
		}
	}
	viewData, err := orm.FindUIDefaultView(ctx, model, viewType)
	if err != nil {
		return nil, workspaceViewNotFoundError(model, []string{viewType}, err)
	}
	arch := strings.TrimSpace(orm.AsString(viewData["arch"]))
	if arch == "" {
		return nil, fmt.Errorf("view arch is empty")
	}
	parsedView, err := parser.ParseViewFromArch(arch)
	if err != nil {
		return nil, fmt.Errorf("parse view arch: %w", err)
	}
	return &resolvedWorkspaceView{
		viewData:     viewData,
		selectedMode: viewType,
		view:         parsedView,
		targetModel:  model,
	}, nil
}

func workspaceViewModeCandidates(r *http.Request, actionData map[string]interface{}) []string {
	modes := splitViewModes(strings.TrimSpace(orm.AsString(actionData["view_mode"])))
	if len(modes) == 0 {
		modes = []string{workspaceViewModeList}
	}

	query := r.URL.Query()
	if viewType := strings.TrimSpace(query.Get(workspaceViewTypeParam)); viewType != "" {
		modes = prependViewMode(normalizeViewMode(viewType), modes)
	}
	if recordID := strings.TrimSpace(query.Get(workspaceRecordIDParam)); isNumericRecordID(recordID) {
		modes = prependViewMode(workspaceViewModeForm, modes)
	}
	return modes
}

func prependViewMode(mode string, modes []string) []string {
	if mode == "" {
		return modes
	}
	return append([]string{mode}, modes...)
}

func isNumericRecordID(recordID string) bool {
	_, err := strconv.Atoi(recordID)
	return err == nil
}

func findFirstWorkspaceView(ctx context.Context, targetModel string, modes []string, actionViewID string) (map[string]interface{}, string, error) {
	primaryMode := ""
	if len(modes) > 0 {
		primaryMode = normalizeViewMode(modes[0])
	}

	var lastErr error
	for _, mode := range modes {
		normalizedMode := normalizeViewMode(mode)
		if normalizedMode == "" {
			continue
		}

		viewData, err := loadUIViewForMode(ctx, targetModel, normalizedMode, actionViewID, primaryMode)
		if err == nil {
			return viewData, normalizedMode, nil
		}
		lastErr = err
	}
	return nil, "", workspaceViewNotFoundError(targetModel, modes, lastErr)
}

func loadUIViewForMode(ctx context.Context, targetModel, mode, actionViewID, primaryMode string) (map[string]interface{}, error) {
	var viewData map[string]interface{}
	var err error

	if actionViewID != "" && mode == primaryMode {
		viewData, err = orm.FindUIViewByName(ctx, targetModel, mode, actionViewID)
	}
	if viewData == nil {
		viewData, err = orm.FindUIDefaultView(ctx, targetModel, mode)
	}
	return viewData, err
}

func workspaceViewNotFoundError(targetModel string, modes []string, lastErr error) error {
	message := fmt.Sprintf("No view for model %s (tried modes: %s)", targetModel, strings.Join(modes, ", "))
	if lastErr != nil && lastErr != sql.ErrNoRows {
		message = fmt.Sprintf("%s: %v", message, lastErr)
	}
	return fmt.Errorf("%s", message)
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
		ViewTabs: render.WorkspaceViewTabs(ctx, resolved.targetModel, req.actionID, req.menuID, resolved.selectedMode, req.recordID),
	}
	appendPageFlashesToViewRecord(r, w, viewRecord)
	appendQueryFlashesToViewRecord(r, viewRecord)

	if recordID, ok := parsePositiveRecordID(req.recordID); ok {
		viewRecord.RecordID = recordID
	}

	if err := loadViewModeData(ctx, viewRecord, resolved, actionData, req); err != nil {
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

func loadViewModeData(ctx context.Context, viewRecord *render.ViewRecordData, resolved *resolvedWorkspaceView, actionData map[string]interface{}, req workspaceRequest) error {
	switch resolved.selectedMode {
	case workspaceViewModeForm:
		return loadWorkspaceFormData(ctx, viewRecord, resolved.targetModel, req.recordID, actionData)
	case workspaceViewModeList:
		viewRecord.ListSearchQuery = req.listSearch
		viewRecord.ListSearchURL = workspaceListSearchURL(req)
		viewRecord.ListSort = req.listSort
		viewRecord.ListOffset = req.listOffset
		viewRecord.ListFilter = req.listFilter
		return loadWorkspaceListData(ctx, viewRecord, resolved.targetModel, actionData, resolved.view, req)
	case workspaceViewModeKanban:
		viewRecord.ListSearchQuery = req.listSearch
		viewRecord.ListFilter = req.listFilter
		return loadWorkspaceKanbanData(ctx, viewRecord, resolved, actionData, req)
	case workspaceViewModePivot:
		return loadWorkspacePivotData(ctx, viewRecord, resolved, actionData, req)
	case workspaceViewModeGraph, workspaceViewModeCalendar:
		viewRecord.ListSearchQuery = req.listSearch
		viewRecord.ListFilter = req.listFilter
		return loadWorkspaceCollectionData(ctx, viewRecord, resolved.targetModel, actionData, resolved.view, req, maxWorkspaceListRows)
	default:
		return nil
	}
}

func loadWorkspaceFormData(ctx context.Context, viewRecord *render.ViewRecordData, targetModel, recordIDRaw string, actionData map[string]interface{}) error {
	if recordIDRaw == "" {
		if defaults := actionDefaultFieldValues(actionData); len(defaults) > 0 {
			viewRecord.Record = defaults
		}
		return nil
	}

	record, err := loadWorkspaceFormRecord(ctx, targetModel, recordIDRaw)
	if err != nil {
		return err
	}
	viewRecord.Record = record
	return nil
}

func loadWorkspaceFormRecord(ctx context.Context, targetModel, recordIDRaw string) (map[string]interface{}, error) {
	recordID, err := strconv.Atoi(recordIDRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}

	record, err := orm.SearchOne(ctx, targetModel, map[string]interface{}{"id": recordID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("record %d not found", recordID)
		}
		return nil, fmt.Errorf("load record: %w", err)
	}

	if targetModel == coreUserModel {
		if companyIDs, err := orm.UserCompanyIDsForUser(ctx, recordID); err == nil {
			record[companyIDsField] = companyIDs
		}
	}
	return record, nil
}

func loadWorkspaceListData(ctx context.Context, viewRecord *render.ViewRecordData, targetModel string, actionData map[string]interface{}, view *parser.View, req workspaceRequest) error {
	searchView := loadSearchView(ctx, targetModel)
	domain := workspaceListDomain(ctx, actionData, view, searchView, req.listSearch, req.listFilter)
	orderBy := orderByFromSortParam(req.listSort)
	rows, err := orm.SearchPage(ctx, targetModel, domain, workspaceListPageSize, req.listOffset, orderBy)
	if err != nil {
		return fmt.Errorf("list load: %w", err)
	}
	total, err := orm.SearchCount(ctx, targetModel, domain)
	if err != nil {
		return fmt.Errorf("list count: %w", err)
	}
	viewRecord.ListRows = rows
	viewRecord.ListTotal = total
	return nil
}

func loadWorkspaceCollectionData(ctx context.Context, viewRecord *render.ViewRecordData, targetModel string, actionData map[string]interface{}, view *parser.View, req workspaceRequest, rowLimit int) error {
	searchView := loadSearchView(ctx, targetModel)
	domain := workspaceListDomain(ctx, actionData, view, searchView, req.listSearch, req.listFilter)
	rows, err := orm.SearchLimit(ctx, targetModel, domain, rowLimit)
	if err != nil {
		return fmt.Errorf("collection load: %w", err)
	}
	viewRecord.ListRows = rows
	return nil
}

func loadWorkspaceKanbanData(ctx context.Context, viewRecord *render.ViewRecordData, resolved *resolvedWorkspaceView, actionData map[string]interface{}, req workspaceRequest) error {
	searchView := loadSearchView(ctx, resolved.targetModel)
	domain := workspaceListDomain(ctx, actionData, resolved.view, searchView, req.listSearch, req.listFilter)
	rows, err := orm.SearchLimit(ctx, resolved.targetModel, domain, maxWorkspaceKanbanRows)
	if err != nil {
		return fmt.Errorf("kanban load: %w", err)
	}

	viewRecord.ListRows = rows
	viewRecord.KanbanModel = resolved.targetModel
	if columns, groupField, draggable := swcmeta.BuildKanbanColumns(ctx, resolved.view, rows); groupField != "" {
		viewRecord.KanbanColumns = nil
		viewRecord.KanbanGroupField = groupField
		viewRecord.KanbanDraggable = draggable
		for _, c := range columns {
			viewRecord.KanbanColumns = append(viewRecord.KanbanColumns, render.KanbanColumn{
				Value: c.Value, Label: c.Label, Sequence: c.Sequence,
				Color: c.Color, Fold: c.Fold, Records: c.Records,
			})
		}
	}
	return nil
}

func loadWorkspacePivotData(ctx context.Context, viewRecord *render.ViewRecordData, resolved *resolvedWorkspaceView, actionData map[string]interface{}, req workspaceRequest) error {
	searchView := loadSearchView(ctx, resolved.targetModel)
	domain := workspaceListDomain(ctx, actionData, resolved.view, searchView, req.listSearch, req.listFilter)
	rows, err := orm.SearchLimit(ctx, resolved.targetModel, domain, maxWorkspaceListRows)
	if err != nil {
		return fmt.Errorf("pivot load: %w", err)
	}
	if pivot := swcmeta.BuildPivotData(resolved.view, rows); pivot != nil {
		viewRecord.Pivot = &render.PivotData{
			RowLabels: pivot.RowLabels, ColLabels: pivot.ColLabels,
			Values: pivot.Values, MeasureLabel: pivot.MeasureLabel,
		}
	}
	return nil
}

func searchWorkspaceRowsWithSearch(ctx context.Context, targetModel string, actionData map[string]interface{}, view *parser.View, searchQuery string, rowLimit int) ([]map[string]interface{}, error) {
	domain := workspaceListDomain(ctx, actionData, view, nil, searchQuery, "")
	return orm.SearchLimit(ctx, targetModel, domain, rowLimit)
}

func searchWorkspaceRows(ctx context.Context, targetModel string, actionData map[string]interface{}, rowLimit int) ([]map[string]interface{}, error) {
	return searchWorkspaceRowsWithSearch(ctx, targetModel, actionData, nil, "", rowLimit)
}

func loadSearchView(ctx context.Context, model string) *parser.View {
	if orm.DB == nil {
		return nil
	}
	viewData, err := orm.FindUIDefaultView(ctx, model, "search")
	if err != nil || viewData == nil {
		return nil
	}
	arch := strings.TrimSpace(orm.AsString(viewData["arch"]))
	if arch == "" {
		return nil
	}
	parsed, err := parser.ParseViewFromArch(arch)
	if err != nil {
		return nil
	}
	return parsed
}

func orderByFromSortParam(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return ""
	}
	if strings.HasPrefix(sort, "-") {
		field := strings.TrimPrefix(sort, "-")
		if field == "" {
			return ""
		}
		return field + " DESC"
	}
	return sort
}

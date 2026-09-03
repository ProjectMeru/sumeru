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
	"sumeru/core/orm"
)

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
		listDomain:  strings.TrimSpace(query.Get(workspaceDomainParam)),
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

func actionViewModesForTabs(actionData map[string]interface{}) []string {
	if actionData == nil || len(actionData) == 0 {
		return nil
	}
	modes := splitViewModes(strings.TrimSpace(orm.AsString(actionData["view_mode"])))
	if len(modes) == 0 {
		return []string{workspaceViewModeList}
	}
	out := make([]string, 0, len(modes))
	for _, mode := range modes {
		if normalized := normalizeViewMode(mode); normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}

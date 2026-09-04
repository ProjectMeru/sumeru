package swcmeta

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

var listMode = "list"
var kanbanMode = "kanban"
var graphMode = "graph"
var calendarMode = "calendar"
var pivotMode = "pivot"
var ganttMode = "gantt"
var mapMode = "map"
var cohortMode = "cohort"

// ViewRecordInput is workspace ORM data passed from the web layer (no render import).
type ViewRecordInput struct {
	ActionID         int
	ResModel         string
	RecordID         int
	FormEditing      bool
	CSRFToken        string
	FormBaseQuery    string
	ListSearchQuery  string
	ListSearchURL    string
	ListTotal        int
	ListSort         string
	ListOffset       int
	ListFilter       string
	ListDomain       string
	ListGroupBy      string
	ListSections     []ListSectionMeta
	Record           map[string]interface{}
	ListRows         []map[string]interface{}
	KanbanColumns    []KanbanColumn
	KanbanGroupField string
	KanbanDraggable  bool
	Pivot            *PivotMeta
	ViewTabs         []ViewTab
	Breadcrumbs      []Breadcrumb
	Defaults         map[string]interface{}
	IframeURL        string
}

// BuildWorkspacePayload serializes loaded workspace data for SWC.
func BuildWorkspacePayload(
	ctx context.Context,
	view *parser.View,
	selectedMode string,
	rec ViewRecordInput,
	reqMenuID string,
) WorkspacePayload {
	arch := SerializeViewForUser(ctx, view)
	arch.Type = selectedMode

	if len(rec.KanbanColumns) > 0 || rec.KanbanGroupField != "" {
		quick := false
		columnsPerRow := 4
		if view != nil {
			quick = view.KanbanQuickCreate()
			columnsPerRow = view.KanbanColumnsPerRow()
		}
		arch.Kanban = &KanbanMeta{
			GroupField:    rec.KanbanGroupField,
			Draggable:     rec.KanbanDraggable,
			QuickCreate:   quick,
			ColumnsPerRow: columnsPerRow,
			Columns:       rec.KanbanColumns,
		}
	}
	if rec.Pivot != nil {
		arch.Pivot = rec.Pivot
	}
	if selectedMode == listMode || selectedMode == kanbanMode || selectedMode == graphMode || selectedMode == calendarMode || selectedMode == pivotMode || selectedMode == ganttMode || selectedMode == mapMode || selectedMode == cohortMode {
		if arch.Search == nil {
			arch.Search = loadSearchMeta(ctx, rec.ResModel)
		}
	}

	payload := WorkspacePayload{
		ActionID:      rec.ActionID,
		MenuID:        reqMenuID,
		ViewType:      selectedMode,
		Model:         rec.ResModel,
		RecordID:      rec.RecordID,
		FormEdit:      rec.FormEditing,
		CSRFToken:     rec.CSRFToken,
		Arch:          arch,
		ListSearch:    rec.ListSearchQuery,
		ListSearchURL: rec.ListSearchURL,
		ListTotal:     rec.ListTotal,
		ListSort:      rec.ListSort,
		ListOffset:    rec.ListOffset,
		ListFilter:    rec.ListFilter,
		ListDomain:    rec.ListDomain,
		ListGroupBy:   rec.ListGroupBy,
		ListSections:  rec.ListSections,
		FormBaseQuery: rec.FormBaseQuery,
		ViewTabs:      rec.ViewTabs,
		Breadcrumbs:   rec.Breadcrumbs,
		Defaults:      rec.Defaults,
		IframeURL:     rec.IframeURL,
	}

	if selectedMode == listMode || selectedMode == kanbanMode || selectedMode == graphMode || selectedMode == calendarMode || selectedMode == pivotMode || selectedMode == ganttMode || selectedMode == mapMode || selectedMode == cohortMode {
		payload.Favorites = LoadSavedSearches(ctx, rec.ActionID, rec.ResModel)
	}

	if rec.Record != nil {
		enrichMany2OneNames(ctx, rec.ResModel, []map[string]interface{}{rec.Record})
		payload.Record = redactCopy(ctx, rec.ResModel, rec.Record)
	}
	if len(rec.ListRows) > 0 {
		enrichMany2OneNames(ctx, rec.ResModel, rec.ListRows)
		payload.Records = redactRows(ctx, rec.ResModel, rec.ListRows)
	}
	if len(rec.ListSections) > 0 {
		for i := range rec.ListSections {
			if len(rec.ListSections[i].Records) > 0 {
				enrichMany2OneNames(ctx, rec.ResModel, rec.ListSections[i].Records)
				rec.ListSections[i].Records = redactRows(ctx, rec.ResModel, rec.ListSections[i].Records)
			}
		}
		payload.ListSections = rec.ListSections
	}
	return payload
}

// BuildIframeWorkspacePayload serializes a URL action for SWC iframe view.
func BuildIframeWorkspacePayload(actionID int, menuID, iframeURL, title string) WorkspacePayload {
	iframeURL = strings.TrimSpace(iframeURL)
	if title == "" {
		title = "Report"
	}
	return WorkspacePayload{
		ActionID: actionID,
		MenuID:   menuID,
		ViewType: "iframe",
		Model:    "sys.action.url",
		Arch: ViewArch{
			Type:  "iframe",
			Model: "sys.action.url",
			Title: title,
		},
		IframeURL: iframeURL,
	}
}

func redactCopy(ctx context.Context, model string, rec map[string]interface{}) map[string]interface{} {
	copy := map[string]interface{}{}
	for k, v := range rec {
		copy[k] = v
	}
	uid := orm.UIDFromContext(ctx)
	orm.RedactRecordForRead(ctx, uid, model, copy)
	return copy
}

func redactRows(ctx context.Context, model string, rows []map[string]interface{}) []map[string]interface{} {
	uid := orm.UIDFromContext(ctx)
	out := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		copy := map[string]interface{}{}
		for k, v := range row {
			copy[k] = v
		}
		orm.RedactRecordForRead(ctx, uid, model, copy)
		out[i] = copy
	}
	return out
}

func loadSearchMeta(ctx context.Context, model string) *SearchMeta {
	model = strings.TrimSpace(model)
	if model == "" || orm.DB == nil {
		return BuildSearchMeta(ctx, model, nil)
	}
	viewData, err := orm.FindUIDefaultView(ctx, model, "search")
	if err != nil || viewData == nil {
		return BuildSearchMeta(ctx, model, nil)
	}
	arch := strings.TrimSpace(orm.AsString(viewData["arch"]))
	if arch == "" {
		return BuildSearchMeta(ctx, model, nil)
	}
	parsed, err := parser.ParseViewFromArch(arch)
	if err != nil {
		return BuildSearchMeta(ctx, model, nil)
	}
	return BuildSearchMeta(ctx, model, parsed)
}

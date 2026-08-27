package swcmeta

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

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
	Record           map[string]interface{}
	ListRows         []map[string]interface{}
	KanbanColumns    []KanbanColumn
	KanbanGroupField string
	KanbanDraggable  bool
	Pivot            *PivotMeta
	ViewTabs         []ViewTab
	Breadcrumbs      []Breadcrumb
	Defaults         map[string]interface{}
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
		if view != nil {
			quick = view.KanbanQuickCreate()
		}
		arch.Kanban = &KanbanMeta{
			GroupField:  rec.KanbanGroupField,
			Draggable:   rec.KanbanDraggable,
			QuickCreate: quick,
			Columns:     rec.KanbanColumns,
		}
	}
	if rec.Pivot != nil {
		arch.Pivot = rec.Pivot
	}
	if selectedMode == "list" || selectedMode == "kanban" || selectedMode == "graph" || selectedMode == "calendar" || selectedMode == "pivot" {
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
		FormBaseQuery: rec.FormBaseQuery,
		ViewTabs:      rec.ViewTabs,
		Breadcrumbs:   rec.Breadcrumbs,
		Defaults:      rec.Defaults,
	}

	if rec.Record != nil {
		payload.Record = redactCopy(ctx, rec.ResModel, rec.Record)
	}
	if len(rec.ListRows) > 0 {
		payload.Records = redactRows(ctx, rec.ResModel, rec.ListRows)
	}
	return payload
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
	return serializeSearch(parsed)
}

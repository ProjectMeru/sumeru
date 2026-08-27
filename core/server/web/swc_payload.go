package web

import (
	"context"
	"strconv"

	"sumeru/core/engine/render"
	"sumeru/core/engine/swcmeta"
)

func buildSwcWorkspacePayload(
	ctx context.Context,
	resolved *resolvedWorkspaceView,
	req workspaceRequest,
	viewRecord *render.ViewRecordData,
	actionData map[string]interface{},
) swcmeta.WorkspacePayload {
	viewBC := render.HumanViewBreadcrumb(resolved.view.Model, resolved.selectedMode)
	crumbs := render.BuildWorkspaceBreadcrumbs(ctx, render.BreadcrumbInput{
		ActiveMenuID:  req.menuID,
		ViewType:      resolved.selectedMode,
		Title:         viewBC,
		FormBaseQuery: viewRecord.FormBaseQuery,
		Record:        viewRecord.Record,
		RecordID:      viewRecord.RecordID,
	})
	tabs := render.WorkspaceViewTabs(ctx, viewRecord.ResModel, viewRecord.ActionID, req.menuID, resolved.selectedMode, recordIDStr(viewRecord.RecordID))

	input := swcmeta.ViewRecordInput{
		ActionID:         viewRecord.ActionID,
		ResModel:         viewRecord.ResModel,
		RecordID:         viewRecord.RecordID,
		FormEditing:      viewRecord.FormEditing,
		CSRFToken:        viewRecord.CSRFToken,
		FormBaseQuery:    viewRecord.FormBaseQuery,
		ListSearchQuery:  viewRecord.ListSearchQuery,
		ListSearchURL:    viewRecord.ListSearchURL,
		ListTotal:        viewRecord.ListTotal,
		ListSort:         viewRecord.ListSort,
		ListOffset:       viewRecord.ListOffset,
		ListFilter:       viewRecord.ListFilter,
		ListDomain:       viewRecord.ListDomain,
		ListGroupBy:      viewRecord.ListGroupBy,
		Record:           viewRecord.Record,
		ListRows:         viewRecord.ListRows,
		KanbanGroupField: viewRecord.KanbanGroupField,
		KanbanDraggable:  viewRecord.KanbanDraggable,
		ViewTabs:         serializeSwcViewTabs(tabs),
		Breadcrumbs:      serializeSwcBreadcrumbs(crumbs),
		Defaults:         actionDefaultFieldValues(actionData),
	}
	for _, c := range viewRecord.KanbanColumns {
		input.KanbanColumns = append(input.KanbanColumns, swcmeta.KanbanColumn{
			Value: c.Value, Label: c.Label, Sequence: c.Sequence,
			Color: c.Color, Fold: c.Fold, Records: c.Records,
		})
	}
	if viewRecord.Pivot != nil {
		input.Pivot = &swcmeta.PivotMeta{
			RowLabels: viewRecord.Pivot.RowLabels, ColLabels: viewRecord.Pivot.ColLabels,
			Values: viewRecord.Pivot.Values, MeasureLabel: viewRecord.Pivot.MeasureLabel,
		}
	}
	return swcmeta.BuildWorkspacePayload(ctx, resolved.view, resolved.selectedMode, input, req.menuID)
}

func recordIDStr(id int) string {
	if id <= 0 {
		return ""
	}
	return strconv.Itoa(id)
}

func serializeSwcViewTabs(tabs []render.ViewSwitchTab) []swcmeta.ViewTab {
	out := make([]swcmeta.ViewTab, 0, len(tabs))
	for _, t := range tabs {
		out = append(out, swcmeta.ViewTab{Mode: t.Mode, Label: t.Label, Href: t.Href, Active: t.Active})
	}
	return out
}

func serializeSwcBreadcrumbs(items []render.BreadcrumbItem) []swcmeta.Breadcrumb {
	out := make([]swcmeta.Breadcrumb, 0, len(items))
	for _, b := range items {
		out = append(out, swcmeta.Breadcrumb{Label: b.Label, Href: b.Href})
	}
	return out
}

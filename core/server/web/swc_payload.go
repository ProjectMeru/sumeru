package web

import (
	"context"

	"sumeru/core/engine/render"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
)

func buildSwcWorkspacePayload(
	ctx context.Context,
	resolved *resolvedWorkspaceView,
	req workspaceRequest,
	viewRecord *render.ViewRecordData,
	actionData map[string]interface{},
) swcmeta.WorkspacePayload {
	crumbMenuID := req.menuID
	if actionTarget(actionData) == "main" {
		crumbMenuID = ""
	}
	crumbs := render.BuildWorkspaceBreadcrumbs(ctx, render.BreadcrumbInput{
		ActiveMenuID:  crumbMenuID,
		ResModel:      resolved.view.Model,
		ViewType:      resolved.selectedMode,
		FormBaseQuery: viewRecord.FormBaseQuery,
		Record:        viewRecord.Record,
		RecordID:      viewRecord.RecordID,
	})

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
		ListSections:     serializeListSections(viewRecord.ListSections),
		Record:           viewRecord.Record,
		ListRows:         viewRecord.ListRows,
		KanbanGroupField: viewRecord.KanbanGroupField,
		KanbanDraggable:  viewRecord.KanbanDraggable,
		ViewTabs:         serializeSwcViewTabs(viewRecord.ViewTabs),
		Breadcrumbs:      serializeSwcBreadcrumbs(crumbs),
		Defaults:         actionDefaultFieldValues(actionData),
	}
	for _, c := range viewRecord.KanbanColumns {
		input.KanbanColumns = append(input.KanbanColumns, swcmeta.KanbanColumn{
			Value: c.Value, Label: c.Label, Sequence: c.Sequence,
			Color: c.Color, RottingDays: c.RottingDays, Fold: c.Fold, Records: c.Records,
		})
	}
	if viewRecord.Pivot != nil {
		input.Pivot = &swcmeta.PivotMeta{
			RowLabels: viewRecord.Pivot.RowLabels, ColLabels: viewRecord.Pivot.ColLabels,
			Values: viewRecord.Pivot.Values, MeasureLabel: viewRecord.Pivot.MeasureLabel,
		}
	}
	payload := swcmeta.BuildWorkspacePayload(
		ctx,
		resolved.view,
		resolved.selectedMode,
		input,
		req.menuID,
		actionViewModesForTabs(actionData),
	)
	if viewRecord.ResModel == coreUserModel && resolved.selectedMode == "form" {
		if meta, err := orm.BuildUserSecurityMeta(ctx, orm.SecurityUID(ctx), viewRecord.RecordID); err == nil {
			us := swcmeta.UserSecurityFromORM(meta)
			payload.UserSecurity = &us
		}
	}
	target := actionTarget(actionData)
	payload.Target = target
	payload.Fullscreen = target == "fullscreen"
	applyActivityRecordFlags(ctx, &payload, resolved.selectedMode, viewRecord.RecordID, viewRecord.ResModel)
	return payload
}

func applyActivityRecordFlags(ctx context.Context, payload *swcmeta.WorkspacePayload, viewMode string, recordID int, model string) {
	_ = viewMode
	if payload == nil || model == "" {
		return
	}
	uid := orm.SecurityUID(ctx)
	if err := orm.CheckModelAccess(ctx, uid, model, "read"); err != nil {
		return
	}
	if !orm.RecordAuditLogAvailable(ctx) {
		return
	}
	payload.RecordLogEligible = true
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

func serializeListSections(sections []render.ListSection) []swcmeta.ListSectionMeta {
	if 0 == len(sections) {
		return nil
	}
	out := make([]swcmeta.ListSectionMeta, 0, len(sections))
	for _, s := range sections {
		out = append(out, swcmeta.ListSectionMeta{
			Label:   s.Label,
			Value:   s.Value,
			Count:   s.Count,
			Records: s.Rows,
		})
	}
	return out
}

package render

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

// WorkspaceTabsInput configures workspace view-mode tab links.
type WorkspaceTabsInput struct {
	ResModel     string
	ActionID     int
	MenuID       string
	SelectedMode string
	RecordID     string
	ViewModes    []string
}

// WorkspaceViewTabs builds URLs for each action view_mode that has a default sys.view for resModel.
// in.ViewModes comes from sys.action.window view_mode (e.g. map, list, form). When empty, every
// known workspace mode is considered (direct ?model= opens without an action).
func WorkspaceViewTabs(ctx context.Context, in WorkspaceTabsInput) []ViewSwitchTab {
	order := []struct {
		mode  string
		label string
	}{
		{ViewModeKanban, "Kanban"},
		{ViewModeList, "List"},
		{ViewModeForm, "Form"},
		{ViewModeGraph, "Graph"},
		{ViewModePivot, "Pivot"},
		{ViewModeCalendar, "Calendar"},
		{ViewModeGantt, "Gantt"},
		{ViewModeMap, "Map"},
		{ViewModeCohort, "Cohort"},
		{ViewModeHierarchy, "Hierarchy"},
		{ViewModeActivity, "Activity"},
	}
	sel := strings.ToLower(strings.TrimSpace(in.SelectedMode))
	menuID := strings.TrimSpace(in.MenuID)
	recID := strings.TrimSpace(in.RecordID)
	allowed := viewModeFilterSet(in.ViewModes)

	var out []ViewSwitchTab
	for _, o := range order {
		if allowed != nil {
			if _, ok := allowed[o.mode]; !ok {
				continue
			}
		}
		if _, err := orm.FindUIDefaultView(ctx, in.ResModel, o.mode); err != nil {
			continue
		}
		q := WorkspaceQuery{ActionID: in.ActionID, MenuID: menuID, ViewType: o.mode}
		if o.mode == ViewModeForm && recID != "" {
			q.RecordID = recID
		}
		out = append(out, ViewSwitchTab{
			Label:  o.label,
			Href:   WorkspaceURL(q),
			Mode:   o.mode,
			Active: sel == o.mode,
		})
	}
	return out
}

// viewModeFilterSet returns nil when all modes should be considered; otherwise a set of allowed modes.
func viewModeFilterSet(modes []string) map[string]struct{} {
	if len(modes) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(modes))
	for _, mode := range modes {
		mode = strings.ToLower(strings.TrimSpace(mode))
		if mode != "" {
			out[mode] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

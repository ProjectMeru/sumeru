package render

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

// WorkspaceViewTabs builds URLs for each action view_mode that has a default sys.view for resModel.
// viewModes comes from sys.action.window view_mode (e.g. map, list, form). When empty, every
// known workspace mode is considered (direct ?model= opens without an action).
func WorkspaceViewTabs(ctx context.Context, resModel string, actionID int, menuID, selectedMode, recordID string, viewModes []string) []ViewSwitchTab {
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
	}
	sel := strings.ToLower(strings.TrimSpace(selectedMode))
	menuID = strings.TrimSpace(menuID)
	recID := strings.TrimSpace(recordID)
	allowed := viewModeFilterSet(viewModes)

	var out []ViewSwitchTab
	for _, o := range order {
		if allowed != nil {
			if _, ok := allowed[o.mode]; !ok {
				continue
			}
		}
		if _, err := orm.FindUIDefaultView(ctx, resModel, o.mode); err != nil {
			continue
		}
		q := WorkspaceQuery{ActionID: actionID, MenuID: menuID, ViewType: o.mode}
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

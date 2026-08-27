package render

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

// WorkspaceViewTabs builds URLs for each view mode that has a default sys.view for resModel.
func WorkspaceViewTabs(ctx context.Context, resModel string, actionID int, menuID, selectedMode, recordID string) []ViewSwitchTab {
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
	}
	sel := strings.ToLower(strings.TrimSpace(selectedMode))
	menuID = strings.TrimSpace(menuID)
	recID := strings.TrimSpace(recordID)

	var out []ViewSwitchTab
	for _, o := range order {
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

package render

import (
	"strings"
)

const (
	labelForm           = "Form"
	breadcrumbSeparator = " · "
)

// Display names for technical models (list/collection context).
var uiModelNames = map[string]string{
	"core.company": "Companies",
	"core.user":    "Users",
	"core.group":   "Groups",
	"core.module":  "Modules",
	"core.menu":    "Menus",
	"core.action":  "Actions",
	"core.view":    "Views",
	"core.rule":    "Access Rules",
	"core.cron":    "Scheduled Actions",
	"core.report":  "Reports",
}

// Breadcrumb labels when list vs record context differs.
var modelBreadcrumbLabels = map[string]struct {
	list, singular string
}{
	"core.company": {list: "Companies", singular: "Company"},
	"core.user":    {list: "Users", singular: "User"},
}

// View-mode suffix appended after breadcrumbSeparator (list/form use base only).
var viewModeSuffix = map[string]string{
	ViewModeKanban:    "Kanban",
	ViewModePivot:     "Pivot",
	ViewModeGraph:     "Graph",
	ViewModeCalendar:  "Calendar",
	ViewModeGantt:     "Gantt",
	ViewModeMap:       "Map",
	ViewModeCohort:    "Cohort",
	ViewModeHierarchy: "Hierarchy",
	ViewModeActivity:  "Activity",
	ViewModeSearch:    "Search",
}

// UIModelName returns a human label for a technical model name.
func UIModelName(technicalModel string) string {
	model := strings.TrimSpace(technicalModel)
	if model == "" {
		return ""
	}
	if name, ok := uiModelNames[model]; ok {
		return name
	}
	parts := strings.Split(model, ".")
	if len(parts) == 0 {
		return model
	}
	last := parts[len(parts)-1]
	return strings.ToUpper(last[:1]) + last[1:]
}

// HumanViewBreadcrumb builds a short breadcrumb label for workspace chrome.
func HumanViewBreadcrumb(technicalModel, viewType string) string {
	model := strings.TrimSpace(technicalModel)
	vt := strings.TrimSpace(viewType)
	if labels, ok := modelBreadcrumbLabels[model]; ok {
		if vt == ViewModeList {
			return labels.list
		}
		return labels.singular
	}
	return viewTypeLabel(UIModelName(model), vt)
}

func viewTypeLabel(base, viewType string) string {
	switch viewType {
	case ViewModeList:
		return base
	case ViewModeForm:
		if base != "" {
			return base
		}
		return labelForm
	default:
		if suffix, ok := viewModeSuffix[viewType]; ok {
			return base + breadcrumbSeparator + suffix
		}
		return base
	}
}

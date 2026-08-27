package render

import (
	"strings"
	"unicode"
)

// UIModelName maps technical model names to end-user labels for the UI.
func UIModelName(technicalModel string) string {
	switch strings.TrimSpace(technicalModel) {
	case "core.company":
		return "Companies"
	case "core.user":
		return "Users"
	case "core.partner":
		return "Contacts"
	case "core.currency":
		return "Currencies"
	case "sys.module":
		return "Apps"
	default:
		return humanizeModelTechnical(technicalModel)
	}
}

// HumanViewBreadcrumb returns the workspace breadcrumb segment for a view (no technical model id).
func HumanViewBreadcrumb(technicalModel, viewType string) string {
	switch strings.TrimSpace(technicalModel) {
	case "core.company":
		switch strings.TrimSpace(viewType) {
		case ViewModeList:
			return "Companies"
		default:
			return "Company"
		}
	case "core.user":
		switch strings.TrimSpace(viewType) {
		case ViewModeList:
			return "Users"
		default:
			return "User"
		}
	default:
		return viewTypeLabel(UIModelName(technicalModel), viewType)
	}
}

func viewTypeLabel(base, viewType string) string {
	switch strings.TrimSpace(viewType) {
	case ViewModeList:
		return base
	case ViewModeForm:
		if base != "" {
			return base
		}
		return "Form"
	case ViewModeKanban:
		return base + " · Kanban"
	case ViewModePivot:
		return base + " · Pivot"
	case ViewModeGraph:
		return base + " · Graph"
	case ViewModeCalendar:
		return base + " · Calendar"
	default:
		return base
	}
}

func humanizeModelTechnical(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	parts := strings.Split(model, ".")
	last := parts[len(parts)-1]
	last = strings.ReplaceAll(last, "_", " ")
	rs := []rune(last)
	for i, r := range rs {
		if i == 0 {
			rs[i] = unicode.ToUpper(r)
		} else if i > 0 && rs[i-1] == ' ' {
			rs[i] = unicode.ToUpper(r)
		}
	}
	return string(rs)
}

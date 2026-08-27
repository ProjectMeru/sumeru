package render

import (
	"net/url"
	"strconv"
	"strings"
)

// Workspace URL path and query parameter names shared by web handlers and render.
const (
	WorkspaceRoute         = "/web"
	WorkspaceActionParam   = "action"
	WorkspaceMenuIDParam   = "menu_id"
	WorkspaceViewTypeParam = "view_type"
	WorkspaceRecordIDParam = "id"
	WorkspaceEditParam     = "edit"
	WorkspaceSearchParam   = "q"
	WorkspaceModelParam    = "model"
	WorkspaceFilterParam   = "filter"
	WorkspaceSortParam     = "sort"
	WorkspaceOffsetParam   = "offset"
	WorkspaceGroupByParam  = "groupby"
)

// Workspace view_type values.
const (
	ViewModeList     = "list"
	ViewModeForm     = "form"
	ViewModeKanban   = "kanban"
	ViewModePivot    = "pivot"
	ViewModeGraph    = "graph"
	ViewModeCalendar = "calendar"
	ViewModeSearch   = "search"
)

// WorkspaceQuery is the /web workspace query (empty fields are omitted).
type WorkspaceQuery struct {
	ActionID int
	Action   string
	MenuID   string
	ViewType string
	RecordID string
	Search   string
	Model    string
	Filter   string
	Sort     string
	Offset   string
	GroupBy  string
}

func setWorkspaceQueryString(query url.Values, param, value string) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		query.Set(param, trimmed)
	}
}

func (q WorkspaceQuery) values() url.Values {
	query := url.Values{}
	if q.ActionID > 0 {
		query.Set(WorkspaceActionParam, strconv.Itoa(q.ActionID))
	} else {
		setWorkspaceQueryString(query, WorkspaceActionParam, q.Action)
	}
	setWorkspaceQueryString(query, WorkspaceMenuIDParam, q.MenuID)
	setWorkspaceQueryString(query, WorkspaceViewTypeParam, q.ViewType)
	setWorkspaceQueryString(query, WorkspaceRecordIDParam, q.RecordID)
	setWorkspaceQueryString(query, WorkspaceSearchParam, q.Search)
	setWorkspaceQueryString(query, WorkspaceModelParam, q.Model)
	setWorkspaceQueryString(query, WorkspaceFilterParam, q.Filter)
	setWorkspaceQueryString(query, WorkspaceSortParam, q.Sort)
	setWorkspaceQueryString(query, WorkspaceOffsetParam, q.Offset)
	setWorkspaceQueryString(query, WorkspaceGroupByParam, q.GroupBy)
	return query
}

// WorkspaceQueryString encodes a workspace query with no leading "?".
func WorkspaceQueryString(q WorkspaceQuery) string {
	return q.values().Encode()
}

// WorkspaceURL builds /web or /web?... from a workspace query.
func WorkspaceURL(q WorkspaceQuery) string {
	encoded := WorkspaceQueryString(q)
	if encoded == "" {
		return WorkspaceRoute
	}
	return WorkspaceRoute + "?" + encoded
}

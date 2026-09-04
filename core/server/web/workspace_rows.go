package web

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
)

func loadViewModeData(ctx context.Context, in workspaceLoadInput) error {
	viewRecord := in.ViewRecord
	resolved := in.Resolved
	req := in.Req
	switch resolved.selectedMode {
	case workspaceViewModeForm:
		return loadWorkspaceFormData(ctx, in)
	case workspaceViewModeList:
		viewRecord.ListSearchQuery = req.listSearch
		viewRecord.ListSearchURL = workspaceListSearchURL(req)
		viewRecord.ListSort = req.listSort
		viewRecord.ListOffset = req.listOffset
		viewRecord.ListFilter = req.listFilter
		viewRecord.ListDomain = req.listDomain
		viewRecord.ListGroupBy = req.listGroupBy
		return loadWorkspaceListData(ctx, in)
	case workspaceViewModeKanban:
		viewRecord.ListSearchQuery = req.listSearch
		viewRecord.ListFilter = req.listFilter
		viewRecord.ListDomain = req.listDomain
		viewRecord.ListGroupBy = req.listGroupBy
		return loadWorkspaceKanbanData(ctx, in)
	case workspaceViewModePivot:
		viewRecord.ListSearchQuery = req.listSearch
		viewRecord.ListFilter = req.listFilter
		viewRecord.ListDomain = req.listDomain
		viewRecord.ListGroupBy = req.listGroupBy
		return loadWorkspacePivotData(ctx, in)
	case workspaceViewModeGraph, workspaceViewModeCalendar, workspaceViewModeGantt, workspaceViewModeMap, workspaceViewModeCohort, workspaceViewModeHierarchy, workspaceViewModeActivity:
		viewRecord.ListSearchQuery = req.listSearch
		viewRecord.ListFilter = req.listFilter
		viewRecord.ListDomain = req.listDomain
		viewRecord.ListGroupBy = req.listGroupBy
		return loadWorkspaceCollectionData(ctx, in, maxWorkspaceListRows)
	default:
		return nil
	}
}

func loadWorkspaceFormData(ctx context.Context, in workspaceLoadInput) error {
	viewRecord := in.ViewRecord
	targetModel := in.Resolved.targetModel
	recordIDRaw := in.Req.recordID
	actionData := in.ActionData
	if recordIDRaw == "" {
		if defaults := actionDefaultFieldValues(actionData); len(defaults) > 0 {
			viewRecord.Record = defaults
		}
		return nil
	}

	record, err := loadWorkspaceFormRecord(ctx, targetModel, recordIDRaw)
	if err != nil {
		return err
	}
	viewRecord.Record = record
	return nil
}

func loadWorkspaceFormRecord(ctx context.Context, targetModel, recordIDRaw string) (map[string]interface{}, error) {
	recordID, err := strconv.Atoi(recordIDRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}

	record, err := orm.SearchOne(ctx, targetModel, map[string]interface{}{"id": recordID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("record %d not found", recordID)
		}
		return nil, fmt.Errorf("load record: %w", err)
	}

	if targetModel == coreUserModel {
		if companyIDs, err := orm.UserCompanyIDsForUser(ctx, recordID); err == nil {
			record[companyIDsField] = companyIDs
		}
	}
	return record, nil
}

func loadWorkspaceListData(ctx context.Context, in workspaceLoadInput) error {
	viewRecord := in.ViewRecord
	targetModel := in.Resolved.targetModel
	req := in.Req
	searchView := loadSearchViewForAction(ctx, targetModel, in.ActionData)
	domain := workspaceListDomain(ctx, listDomainInput{
		ActionData:  in.ActionData,
		View:        in.Resolved.view,
		SearchView:  searchView,
		SearchQuery: req.listSearch,
		FilterCSV:   req.listFilter,
		DomainJSON:  req.listDomain,
	})
	orderBy := orderByFromSortParam(req.listSort)
	if gbFields := splitCommaSeparatedValues(req.listGroupBy); len(gbFields) > 0 {
		gbOrder := strings.Join(gbFields, ", ")
		if orderBy == "" {
			orderBy = gbOrder
		} else {
			orderBy = gbOrder + ", " + orderBy
		}
	}
	rows, err := orm.SearchPage(ctx, targetModel, domain, workspaceListPageSize, req.listOffset, orderBy)
	if err != nil {
		return fmt.Errorf("list load: %w", err)
	}
	total, err := orm.SearchCount(ctx, targetModel, domain)
	if err != nil {
		return fmt.Errorf("list count: %w", err)
	}
	viewRecord.ListRows = rows
	viewRecord.ListTotal = total
	if gb := firstGroupByField(req.listGroupBy); gb != "" {
		viewRecord.ListSections = partitionListSections(rows, gb)
	}
	return nil
}

func loadWorkspaceCollectionData(ctx context.Context, in workspaceLoadInput, rowLimit int) error {
	viewRecord := in.ViewRecord
	targetModel := in.Resolved.targetModel
	req := in.Req
	searchView := loadSearchViewForAction(ctx, targetModel, in.ActionData)
	domain := workspaceListDomain(ctx, listDomainInput{
		ActionData:  in.ActionData,
		View:        in.Resolved.view,
		SearchView:  searchView,
		SearchQuery: req.listSearch,
		FilterCSV:   req.listFilter,
		DomainJSON:  req.listDomain,
	})
	rows, err := orm.SearchLimit(ctx, targetModel, domain, rowLimit)
	if err != nil {
		return fmt.Errorf("collection load: %w", err)
	}
	viewRecord.ListRows = rows
	return nil
}

func loadWorkspaceKanbanData(ctx context.Context, in workspaceLoadInput) error {
	viewRecord := in.ViewRecord
	resolved := in.Resolved
	req := in.Req
	searchView := loadSearchViewForAction(ctx, resolved.targetModel, in.ActionData)
	domain := workspaceListDomain(ctx, listDomainInput{
		ActionData:  in.ActionData,
		View:        resolved.view,
		SearchView:  searchView,
		SearchQuery: req.listSearch,
		FilterCSV:   req.listFilter,
		DomainJSON:  req.listDomain,
	})
	rows, err := orm.SearchLimit(ctx, resolved.targetModel, domain, maxWorkspaceKanbanRows)
	if err != nil {
		return fmt.Errorf("kanban load: %w", err)
	}

	viewRecord.ListRows = rows
	viewRecord.KanbanModel = resolved.targetModel
	groupOverride := firstGroupByField(req.listGroupBy)
	if columns, groupField, draggable := swcmeta.BuildKanbanColumns(ctx, resolved.view, rows, groupOverride); groupField != "" {
		viewRecord.KanbanColumns = nil
		viewRecord.KanbanGroupField = groupField
		viewRecord.KanbanDraggable = draggable
		for _, c := range columns {
			viewRecord.KanbanColumns = append(viewRecord.KanbanColumns, render.KanbanColumn{
				Value: c.Value, Label: c.Label, Sequence: c.Sequence,
				Color: c.Color, RottingDays: c.RottingDays, Fold: c.Fold, Records: c.Records,
			})
		}
	}
	return nil
}

func loadWorkspacePivotData(ctx context.Context, in workspaceLoadInput) error {
	viewRecord := in.ViewRecord
	resolved := in.Resolved
	req := in.Req
	searchView := loadSearchViewForAction(ctx, resolved.targetModel, in.ActionData)
	domain := workspaceListDomain(ctx, listDomainInput{
		ActionData:  in.ActionData,
		View:        resolved.view,
		SearchView:  searchView,
		SearchQuery: req.listSearch,
		FilterCSV:   req.listFilter,
		DomainJSON:  req.listDomain,
	})
	rows, err := orm.SearchLimit(ctx, resolved.targetModel, domain, maxWorkspaceListRows)
	if err != nil {
		return fmt.Errorf("pivot load: %w", err)
	}
	if pivot := swcmeta.BuildPivotData(resolved.view, rows); pivot != nil {
		viewRecord.Pivot = &render.PivotData{
			RowLabels: pivot.RowLabels, ColLabels: pivot.ColLabels,
			Values: pivot.Values, MeasureLabel: pivot.MeasureLabel,
		}
	}
	return nil
}

func loadSearchViewForAction(ctx context.Context, model string, actionData map[string]interface{}) *parser.View {
	if searchViewID := actionSearchViewIDFromContext(actionData); searchViewID != "" {
		if view := loadSearchViewByName(ctx, model, searchViewID); view != nil {
			return view
		}
	}
	return loadSearchView(ctx, model)
}

func loadSearchViewByName(ctx context.Context, model, viewName string) *parser.View {
	if orm.DB == nil {
		return nil
	}
	viewData, err := orm.FindUIViewByName(ctx, model, "search", viewName)
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
	return parsed
}

func loadSearchView(ctx context.Context, model string) *parser.View {
	if orm.DB == nil {
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
	return parsed
}

func orderByFromSortParam(sort string) string {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return ""
	}
	if strings.HasPrefix(sort, "-") {
		field := strings.TrimPrefix(sort, "-")
		if field == "" {
			return ""
		}
		return field + " DESC"
	}
	return sort
}

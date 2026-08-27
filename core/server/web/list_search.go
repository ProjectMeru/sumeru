package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

const workspaceListSearchParam = render.WorkspaceSearchParam

func listSearchQuery(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get(workspaceListSearchParam))
}

func listSearchFieldNames(views ...*parser.View) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, view := range views {
		if view == nil {
			continue
		}
		for _, f := range view.Field {
			n := strings.TrimSpace(f.Name)
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			out = append(out, n)
		}
	}
	return out
}

func workspaceListDomain(ctx context.Context, actionData map[string]interface{}, view, searchView *parser.View, searchQuery, filterCSV, domainJSON string) [][]interface{} {
	base := actionListDomain(ctx, actionData)
	model := ""
	if view != nil {
		model = view.Model
	}
	searchFields := listSearchFieldNames(view, searchView)
	if searchQuery != "" && model != "" {
		search := orm.BuildListSearchDomain(model, searchFields, searchQuery)
		base = orm.MergeDomains(base, search)
	}
	uid := orm.SecurityUID(ctx)
	for _, name := range splitCommaSeparatedValues(filterCSV) {
		f := findSearchFilter(searchView, name)
		if f == nil || strings.TrimSpace(f.Domain) == "" {
			continue
		}
		dom, err := orm.ParseDomainJSON(f.Domain)
		if err != nil || len(dom) == 0 {
			continue
		}
		dom = orm.ResolveDomainXMLRefs(ctx, dom)
		dom = orm.SubstituteDomainUID(dom, uid)
		base = orm.MergeDomains(base, dom)
	}
	if strings.TrimSpace(domainJSON) != "" && model != "" {
		dom, err := orm.ParseDomainJSON(domainJSON)
		if err == nil && len(dom) > 0 {
			dom = filterDomainToModel(model, dom)
			dom = orm.ResolveDomainXMLRefs(ctx, dom)
			dom = orm.SubstituteDomainUID(dom, uid)
			base = orm.MergeDomains(base, dom)
		}
	}
	return base
}

func filterDomainToModel(model string, domain [][]interface{}) [][]interface{} {
	if model == "" || len(domain) == 0 {
		return domain
	}
	inst, ok := orm.Registry[model]
	if !ok {
		return nil
	}
	allowed := map[string]struct{}{}
	for _, fd := range inst.Fields() {
		allowed[fd.Name] = struct{}{}
	}
	out := make([][]interface{}, 0, len(domain))
	for _, d := range domain {
		if len(d) != 3 {
			continue
		}
		field, _ := d[0].(string)
		if _, ok := allowed[field]; !ok {
			continue
		}
		out = append(out, d)
	}
	return out
}

func findSearchFilter(searchView *parser.View, name string) *parser.SearchFilter {
	if searchView == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	for i := range searchView.SearchFilter {
		if searchView.SearchFilter[i].Name == name {
			return &searchView.SearchFilter[i]
		}
	}
	return nil
}

func workspaceListSearchURL(req workspaceRequest) string {
	offset := ""
	if req.listOffset > 0 {
		offset = strconv.Itoa(req.listOffset)
	}
	return render.WorkspaceURL(render.WorkspaceQuery{
		ActionID: req.actionID,
		MenuID:   req.menuID,
		ViewType: workspaceViewModeList,
		Search:   req.listSearch,
		Model:    req.model,
		Filter:   req.listFilter,
		Sort:     req.listSort,
		Offset:   offset,
		GroupBy:  req.listGroupBy,
		Domain:   req.listDomain,
	})
}

package web

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/security"
)

const settingsFieldACLRoute = "/web/settings/field-acl"
const settingsFieldACLExportRoute = settingsFieldACLRoute + "/export"

type fieldACLMatrixCell struct {
	FieldName  string
	GroupID    int
	GroupLabel string
	DenyRead   bool
	DenyWrite  bool
}

type fieldACLMatrixRow struct {
	FieldName    string
	SchemaGroups bool
	KernelTags   []string
	Cells        []fieldACLMatrixCell
}

type settingsFieldACLData struct {
	CSRFToken       string
	Model           string
	ModelChoices    []string
	GroupFilter     string
	FieldFilter     string
	ShowAllGroups   bool
	GroupsTruncated bool
	Groups          []aclGroupCol
	Rows            []fieldACLMatrixRow
	DebugAccessHref string
	Flash           render.FlashMessage
}

func registerSettingsFieldACLRoutes() {
	registerSession(http.MethodGet, settingsFieldACLRoute, SettingsFieldACLGetHandler)
	registerSession(http.MethodPost, settingsFieldACLRoute, SettingsFieldACLPostHandler)
	registerSession(http.MethodGet, settingsFieldACLExportRoute, SettingsFieldACLExportHandler)
}

func SettingsFieldACLGetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSystemAdmin(w, r, true) {
		return
	}
	ctx := r.Context()
	rootMenuID, ok := resolveSettingsRootMenuID(w, r, ctx)
	if !ok {
		return
	}
	menuID := resolveSettingsMenuXMLID(ctx, render.MenuFieldAccessMatrixXMLID, rootMenuID)
	q := r.URL.Query()
	model := strings.TrimSpace(q.Get("model"))
	flash, _ := FlashFromQueryMessage(q.Get("msg"))
	page, err := buildFieldACLMatrixPage(ctx, model, q.Get("group_q"), q.Get("field_q"), q.Get("show_all") == "1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	page.CSRFToken = CSRFTokenForRequest(r)
	page.Flash = flash
	page.ModelChoices = listRegistryModelNames()
	renderSettingsFieldACLPage(w, r, page, menuID)
}

func SettingsFieldACLExportHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSystemAdmin(w, r, true) {
		return
	}
	model := strings.TrimSpace(r.URL.Query().Get("model"))
	if model == "" || orm.RegistryModel(model) == nil {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	body, err := exportFieldACLCSV(ctx, model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeCSVAttachment(w, "field_access_"+model+".csv", body)
}

func SettingsFieldACLPostHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSystemAdmin(w, r, true) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	if !ValidateCSRF(r) {
		http.Error(w, "invalid csrf", http.StatusForbidden)
		return
	}
	model := strings.TrimSpace(r.PostForm.Get("model"))
	if model == "" || orm.RegistryModel(model) == nil {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	err := orm.WithElevated(ctx, "settings.field_acl_matrix", func(elevated context.Context) error {
		groups, err := loadACLGroups(elevated, r.PostForm.Get("group_q"))
		if err != nil {
			return err
		}
		display, _ := matrixGroupsForDisplay(groups, r.PostForm.Get("show_all") == "1")
		return applyFieldACLMatrixPost(elevated, model, r.PostForm, display)
	})
	if err != nil {
		redirectWithWebMessage(w, r, fieldACLRedirectURL(model, r.PostForm), "field_acl_failed")
		return
	}
	redirectWithWebMessage(w, r, fieldACLRedirectURL(model, r.PostForm), "field_acl_saved")
}

func fieldACLRedirectURL(model string, form map[string][]string) string {
	u := settingsFieldACLRoute + "?model=" + model
	if gq := strings.TrimSpace(firstFormVal(form, "group_q")); gq != "" {
		u += "&group_q=" + gq
	}
	if fq := strings.TrimSpace(firstFormVal(form, "field_q")); fq != "" {
		u += "&field_q=" + fq
	}
	if formHas(form, "show_all") {
		u += "&show_all=1"
	}
	return u
}

func firstFormVal(form map[string][]string, key string) string {
	if vals, ok := form[key]; ok && len(vals) > 0 {
		return strings.TrimSpace(vals[0])
	}
	return ""
}

func applyFieldACLMatrixPost(ctx context.Context, model string, form map[string][]string, displayGroups []aclGroupCol) error {
	groupIDs := make([]int, 0, len(displayGroups))
	for _, g := range displayGroups {
		groupIDs = append(groupIDs, g.ID)
	}
	fields := listFieldACLFields(model)
	existing, err := loadFieldACLExisting(ctx, model)
	if err != nil {
		return err
	}
	for _, field := range fields {
		for _, gid := range groupIDs {
			key := matrixCellKey(gid, field)
			denyRead := formHas(form, "dr_"+key)
			denyWrite := formHas(form, "dw_"+key)
			rowName := fieldACLMatrixRowName(model, field, gid)
			if !denyRead && !denyWrite {
				if id, ok := existing[rowName]; ok && id > 0 {
					if err := orm.Unlink(ctx, "sys.field.access", id); err != nil {
						return err
					}
				}
				continue
			}
			values := map[string]interface{}{
				"name":       rowName,
				"model":      model,
				"field_name": field,
				"perm_read":  !denyRead,
				"perm_write": !denyWrite,
			}
			if gid > 0 {
				values["group_id"] = gid
			}
			if _, err := orm.Upsert(ctx, orm.RegistryModel("sys.field.access"), values, "name"); err != nil {
				return err
			}
		}
	}
	return nil
}

func fieldACLMatrixRowName(model, field string, groupID int) string {
	if 0 >= groupID {
		return "matrix." + model + "." + field + ".global"
	}
	return fmt.Sprintf("matrix.%s.%s.g%d", model, field, groupID)
}

func buildFieldACLMatrixPage(ctx context.Context, model, groupFilter, fieldFilter string, showAllGroups bool) (settingsFieldACLData, error) {
	out := settingsFieldACLData{
		Model:         model,
		GroupFilter:   groupFilter,
		FieldFilter:   fieldFilter,
		ShowAllGroups: showAllGroups,
	}
	if model == "" {
		return out, nil
	}
	if orm.RegistryModel(model) == nil {
		return out, fmt.Errorf("unknown model %q", model)
	}
	out.DebugAccessHref = "/web/debug/access?model=" + model
	groups, err := loadACLGroups(ctx, groupFilter)
	if err != nil {
		return out, err
	}
	display, truncated := matrixGroupsForDisplay(groups, showAllGroups)
	out.Groups = display
	out.GroupsTruncated = truncated
	existing, err := loadFieldACLExistingByCell(ctx, model)
	if err != nil {
		return out, err
	}
	fields := filterFieldNames(listFieldACLFields(model), fieldFilter)
	for _, field := range fields {
		row := fieldACLMatrixRow{
			FieldName:    field,
			SchemaGroups: fieldHasGroupsAttr(model, field),
			KernelTags:   security.FieldKernelPolicyTags(model, field),
		}
		for _, g := range display {
			key := matrixCellKey(g.ID, field)
			cell := fieldACLMatrixCell{
				FieldName:  field,
				GroupID:    g.ID,
				GroupLabel: g.Label,
			}
			if st, ok := existing[key]; ok {
				cell.DenyRead = !st.read
				cell.DenyWrite = !st.write
			}
			row.Cells = append(row.Cells, cell)
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}

type fieldACLPerm struct {
	read  bool
	write bool
}

func loadFieldACLExistingByCell(ctx context.Context, model string) (map[string]fieldACLPerm, error) {
	rows, err := orm.Search(ctx, "sys.field.access", [][]interface{}{{"model", "=", model}})
	if err != nil {
		return nil, err
	}
	out := map[string]fieldACLPerm{}
	for _, row := range rows {
		field := stringField(row["field_name"])
		if field == "" {
			continue
		}
		gid := intField(row["group_id"])
		key := matrixCellKey(gid, field)
		out[key] = fieldACLPerm{
			read:  boolField(row["perm_read"], true),
			write: boolField(row["perm_write"], true),
		}
	}
	return out, nil
}

func loadFieldACLExisting(ctx context.Context, model string) (map[string]int, error) {
	rows, err := orm.Search(ctx, "sys.field.access", [][]interface{}{{"model", "=", model}})
	if err != nil {
		return nil, err
	}
	out := map[string]int{}
	for _, row := range rows {
		name := stringField(row["name"])
		id := intField(row["id"])
		if name != "" && id > 0 {
			out[name] = id
		}
	}
	return out, nil
}

func exportFieldACLCSV(ctx context.Context, model string) (string, error) {
	rows, err := orm.Search(ctx, "sys.field.access", [][]interface{}{{"model", "=", model}})
	if err != nil {
		return "", err
	}
	sort.Slice(rows, func(i, j int) bool {
		fi := stringField(rows[i]["field_name"])
		fj := stringField(rows[j]["field_name"])
		if fi != fj {
			return fi < fj
		}
		return intField(rows[i]["group_id"]) < intField(rows[j]["group_id"])
	})
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "name", "model", "field_name", "group_id:id", "perm_read", "perm_write"})
	for _, row := range rows {
		gid := intField(row["group_id"])
		groupRef := ""
		if gid > 0 {
			groupRef = fmt.Sprintf("%d", gid)
		}
		_ = w.Write([]string{
			stringField(row["name"]),
			stringField(row["name"]),
			model,
			stringField(row["field_name"]),
			groupRef,
			fmt.Sprintf("%v", boolField(row["perm_read"], true)),
			fmt.Sprintf("%v", boolField(row["perm_write"], true)),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func listFieldACLFields(model string) []string {
	inst := orm.RegistryModel(model)
	if inst == nil {
		return nil
	}
	skip := map[string]bool{"id": true, "create_uid": true, "write_uid": true, "create_date": true, "write_date": true}
	var names []string
	for _, f := range inst.Fields() {
		n := strings.TrimSpace(f.Name)
		if n == "" || skip[n] {
			continue
		}
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func fieldHasGroupsAttr(model, field string) bool {
	inst := orm.RegistryModel(model)
	if inst == nil {
		return false
	}
	for _, f := range inst.Fields() {
		if strings.TrimSpace(f.Name) == field {
			return strings.TrimSpace(f.Groups) != ""
		}
	}
	return false
}

func renderSettingsFieldACLPage(w http.ResponseWriter, r *http.Request, pageData settingsFieldACLData, menuIDStr string) {
	ctx := r.Context()
	renderShellPage(w, r, shellPageOpts{
		Route:         settingsFieldACLRoute,
		InnerTemplate: settingsFieldACLInnerTemplate,
		InnerData:     pageData,
		MenuIDStr:     menuIDStr,
		Page:          buildSettingsFieldACLPageData(ctx, menuIDStr, pageData),
	})
}

func buildSettingsFieldACLPageData(ctx context.Context, menuIDStr string, page settingsFieldACLData) render.PageData {
	crumbs := render.BuildSettingsHubBreadcrumbs(ctx)
	crumbs = append(crumbs, render.BreadcrumbItem{Label: "Field access matrix"})
	pd := render.PageData{
		Title:              "Field access matrix",
		SettingsNavActive:  true,
		ActiveMenuID:       menuIDStr,
		BreadcrumbItems:    crumbs,
		ViewStylesheetURLs: []string{settingsHubStylesheetURL},
		ExtraBodyClasses:   settingsHubBodyClass,
	}
	if page.Flash.Body != "" || page.Flash.Title != "" {
		pd.FlashMessages = []render.FlashMessage{page.Flash}
	}
	return pd
}

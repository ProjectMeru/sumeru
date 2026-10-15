package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

const settingsModelACLRoute = "/web/settings/model-acl"

type modelACLMatrixCell struct {
	GroupID    int
	GroupLabel string
	PermRead   bool
	PermWrite  bool
	PermCreate bool
	PermUnlink bool
}

type settingsModelACLData struct {
	CSRFToken       string
	Model           string
	ModelChoices    []string
	GroupFilter     string
	ShowAllGroups   bool
	GroupsTruncated bool
	Groups          []aclGroupCol
	Cells           []modelACLMatrixCell
	DebugAccessHref string
	Flash           render.FlashMessage
}

func registerSettingsModelACLRoutes() {
	registerSession(http.MethodGet, settingsModelACLRoute, SettingsModelACLGetHandler)
	registerSession(http.MethodPost, settingsModelACLRoute, SettingsModelACLPostHandler)
}

func SettingsModelACLGetHandler(w http.ResponseWriter, r *http.Request) {
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
	menuID := resolveSettingsMenuXMLID(ctx, render.MenuModelAccessMatrixXMLID, rootMenuID)
	q := r.URL.Query()
	model := strings.TrimSpace(q.Get("model"))
	flash, _ := FlashFromQueryMessage(q.Get("msg"))
	page, err := buildModelACLMatrixPage(ctx, model, q.Get("group_q"), q.Get("show_all") == "1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	page.CSRFToken = CSRFTokenForRequest(r)
	page.Flash = flash
	page.ModelChoices = listRegistryModelNames()
	renderSettingsModelACLPage(w, r, page, menuID)
}

func SettingsModelACLPostHandler(w http.ResponseWriter, r *http.Request) {
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
	err := orm.WithElevated(ctx, "settings.model_acl_matrix", func(elevated context.Context) error {
		groups, err := loadACLGroups(elevated, r.PostForm.Get("group_q"))
		if err != nil {
			return err
		}
		display, _ := matrixGroupsForDisplay(groups, r.PostForm.Get("show_all") == "1")
		return applyModelACLMatrixPost(elevated, model, r.PostForm, display)
	})
	if err != nil {
		redirectWithWebMessage(w, r, modelACLRedirectURL(model, r.PostForm), "model_acl_failed")
		return
	}
	redirectWithWebMessage(w, r, modelACLRedirectURL(model, r.PostForm), "model_acl_saved")
}

func modelACLRedirectURL(model string, form map[string][]string) string {
	u := settingsModelACLRoute + "?model=" + model
	if gq := strings.TrimSpace(firstFormVal(form, "group_q")); gq != "" {
		u += "&group_q=" + gq
	}
	if formHas(form, "show_all") {
		u += "&show_all=1"
	}
	return u
}

func modelACLMatrixRowName(model string, groupID int) string {
	if 0 >= groupID {
		return "matrix." + model + ".global"
	}
	return fmt.Sprintf("matrix.%s.g%d", model, groupID)
}

func applyModelACLMatrixPost(ctx context.Context, model string, form map[string][]string, displayGroups []aclGroupCol) error {
	existing, err := loadModelACLExisting(ctx, model)
	if err != nil {
		return err
	}
	for _, g := range displayGroups {
		key := fmt.Sprintf("%d", g.ID)
		read := formHas(form, "pr_"+key)
		write := formHas(form, "pw_"+key)
		create := formHas(form, "pc_"+key)
		unlink := formHas(form, "pu_"+key)
		rowName := modelACLMatrixRowName(model, g.ID)
		if !read && !write && !create && !unlink {
			if id, ok := existing[rowName]; ok && id > 0 {
				if err := orm.Unlink(ctx, "sys.access", id); err != nil {
					return err
				}
			}
			continue
		}
		values := map[string]interface{}{
			"name":        rowName,
			"model":       model,
			"perm_read":   read,
			"perm_write":  write,
			"perm_create": create,
			"perm_unlink": unlink,
		}
		if g.ID > 0 {
			values["group_id"] = g.ID
		}
		if _, err := orm.Upsert(ctx, orm.RegistryModel("sys.access"), values, "name"); err != nil {
			return err
		}
	}
	return nil
}

func loadModelACLExisting(ctx context.Context, model string) (map[string]int, error) {
	rows, err := orm.Search(ctx, "sys.access", [][]interface{}{{"model", "=", model}})
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

func buildModelACLMatrixPage(ctx context.Context, model, groupFilter string, showAllGroups bool) (settingsModelACLData, error) {
	out := settingsModelACLData{
		Model:         model,
		GroupFilter:   groupFilter,
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
	existing, err := loadModelACLPerms(ctx, model)
	if err != nil {
		return out, err
	}
	for _, g := range display {
		key := fmt.Sprintf("%d", g.ID)
		cell := modelACLMatrixCell{GroupID: g.ID, GroupLabel: g.Label}
		if st, ok := existing[key]; ok {
			cell.PermRead = st.read
			cell.PermWrite = st.write
			cell.PermCreate = st.create
			cell.PermUnlink = st.unlink
		}
		out.Cells = append(out.Cells, cell)
	}
	return out, nil
}

type modelACLPerm struct {
	read, write, create, unlink bool
}

func loadModelACLPerms(ctx context.Context, model string) (map[string]modelACLPerm, error) {
	rows, err := orm.Search(ctx, "sys.access", [][]interface{}{{"model", "=", model}})
	if err != nil {
		return nil, err
	}
	out := map[string]modelACLPerm{}
	for _, row := range rows {
		gid := intField(row["group_id"])
		key := fmt.Sprintf("%d", gid)
		out[key] = modelACLPerm{
			read:   boolField(row["perm_read"], false),
			write:  boolField(row["perm_write"], false),
			create: boolField(row["perm_create"], false),
			unlink: boolField(row["perm_unlink"], false),
		}
	}
	return out, nil
}

func renderSettingsModelACLPage(w http.ResponseWriter, r *http.Request, pageData settingsModelACLData, menuIDStr string) {
	ctx := r.Context()
	renderShellPage(w, r, shellPageOpts{
		Route:         settingsModelACLRoute,
		InnerTemplate: settingsModelACLInnerTemplate,
		InnerData:     pageData,
		MenuIDStr:     menuIDStr,
		Page:          buildSettingsModelACLPageData(ctx, menuIDStr, pageData),
	})
}

func buildSettingsModelACLPageData(ctx context.Context, menuIDStr string, page settingsModelACLData) render.PageData {
	crumbs := render.BuildSettingsHubBreadcrumbs(ctx)
	crumbs = append(crumbs, render.BreadcrumbItem{Label: "Model access matrix"})
	pd := render.PageData{
		Title:              "Model access matrix",
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

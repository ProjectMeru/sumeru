package web

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const portalHomeModel = "core.partner"

type portalHomeData struct {
	AppName     string
	LogoURL     string
	Stylesheets []string
	Records     []portalRecordRow
	Error       string
}

type portalRecordRow struct {
	ID   int64
	Name string
}

type portalShareData struct {
	AppName     string
	LogoURL     string
	Stylesheets []string
	Model       string
	Record      map[string]interface{}
	Error       string
}

var (
	portalTemplateOnce sync.Once
	portalHomeTmpl     *template.Template
	portalShareTmpl    *template.Template
	portalTemplateErr  error
)

func PortalHomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !requirePortalUser(w, r) {
		return
	}
	data := portalHomeData{
		AppName:     render.AppDisplayName,
		LogoURL:     render.ShellLogoURL(),
		Stylesheets: assets.PortalStylesheetURLs(),
	}
	if !orm.PortalAllowsModel(portalHomeModel, "read") {
		data.Error = "Portal access not configured"
		writePortalHome(w, data)
		return
	}
	rows, err := orm.SearchLimit(r.Context(), portalHomeModel, nil, 50)
	if err != nil {
		data.Error = "Could not load records"
		writePortalHome(w, data)
		return
	}
	for _, row := range rows {
		id, ok := orm.CoerceInt64(row["id"])
		if !ok {
			continue
		}
		data.Records = append(data.Records, portalRecordRow{
			ID:   id,
			Name: strings.TrimSpace(orm.AsString(row["name"])),
		})
	}
	writePortalHome(w, data)
}

func PortalShareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token := strings.TrimPrefix(r.URL.Path, portalShareRoutePrefix)
	token = strings.Trim(token, "/")
	data := portalShareData{
		AppName:     render.AppDisplayName,
		LogoURL:     render.ShellLogoURL(),
		Stylesheets: assets.PortalStylesheetURLs(),
	}
	parsed, err := parseShareToken(token)
	if err != nil {
		data.Error = "Link invalid or expired"
		writePortalShare(w, data)
		return
	}
	if !orm.PortalAllowsModel(parsed.Model, "read") {
		data.Error = "Access denied"
		writePortalShare(w, data)
		return
	}
	ctx := orm.ContextWithUID(r.Context(), parsed.IssuerUID)
	ctx = orm.ContextWithCompanyID(ctx, orm.ActiveCompanyIDForUser(ctx, parsed.IssuerUID))
	rec, err := orm.SearchOne(ctx, parsed.Model, map[string]interface{}{"id": parsed.ResID})
	if err != nil || rec == nil {
		data.Error = "Record not available"
		writePortalShare(w, data)
		return
	}
	data.Model = parsed.Model
	data.Record = rec
	writePortalShare(w, data)
}

func PortalRecordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !requirePortalUser(w, r) {
		return
	}
	model := strings.TrimSpace(r.URL.Query().Get("model"))
	id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if model == "" || id <= 0 || !orm.PortalAllowsModel(model, "read") {
		http.NotFound(w, r)
		return
	}
	rec, err := orm.SearchOne(r.Context(), model, map[string]interface{}{"id": id})
	if err != nil || rec == nil {
		http.NotFound(w, r)
		return
	}
	data := portalShareData{
		AppName:     render.AppDisplayName,
		LogoURL:     render.ShellLogoURL(),
		Stylesheets: assets.PortalStylesheetURLs(),
		Model:       model,
		Record:      rec,
	}
	writePortalShare(w, data)
}

func loadPortalTemplates() (*template.Template, *template.Template, error) {
	portalTemplateOnce.Do(func() {
		dir := config.AppConfig.TemplatesPath
		portalHomeTmpl, portalTemplateErr = template.ParseFiles(
			filepath.Join(dir, "portal_home.html"),
		)
		if portalTemplateErr != nil {
			return
		}
		portalShareTmpl, portalTemplateErr = template.ParseFiles(
			filepath.Join(dir, "portal_share.html"),
		)
	})
	return portalHomeTmpl, portalShareTmpl, portalTemplateErr
}

func writePortalHome(w http.ResponseWriter, data portalHomeData) {
	homeTmpl, _, err := loadPortalTemplates()
	if err != nil || homeTmpl == nil {
		http.Error(w, "Portal unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = homeTmpl.Execute(w, data)
}

func writePortalShare(w http.ResponseWriter, data portalShareData) {
	_, shareTmpl, err := loadPortalTemplates()
	if err != nil || shareTmpl == nil {
		http.Error(w, "Portal unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = shareTmpl.Execute(w, data)
}

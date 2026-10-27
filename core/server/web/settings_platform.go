package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/mail"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const (
	settingsAuthRoute           = "/web/settings/auth"
	settingsMailRoute           = "/web/settings/mail"
	authPasswordMinLengthKey    = "auth.password_min_length"
	authPasswordMinLengthDefault = "8"
)

func registerSettingsPlatformRoutes() {
	registerSession(http.MethodGet, settingsAuthRoute, SettingsAuthGetHandler)
	registerSession(http.MethodPost, settingsAuthRoute, SettingsAuthPostHandler)
	registerSession(http.MethodGet, settingsMailRoute, SettingsMailGetHandler)
}

type settingsAuthData struct {
	CSRFToken          string
	Flash              render.FlashMessage
	LocalLoginEnabled  bool
	PasswordMinLength  int
}

type settingsMailData struct {
	CSRFToken string
	Flash     render.FlashMessage
	Configured bool
	SMTPHost   string
	SMTPPort   int
	SMTPFrom   string
	SMTPUser   string
}

func SettingsAuthGetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSystemAdmin(w, r, true) {
		return
	}
	ctx := r.Context()
	menuIDStr, ok := resolveSettingsRootMenuID(w, r, ctx)
	if !ok {
		return
	}
	flash, _ := FlashFromQueryMessage(r.URL.Query().Get("msg"))
	renderSettingsAuthPage(w, r, loadSettingsAuthData(ctx, CSRFTokenForRequest(r), flash), menuIDStr)
}

func SettingsAuthPostHandler(w http.ResponseWriter, r *http.Request) {
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
	ctx := r.Context()
	localEnabled := strings.TrimSpace(r.PostForm.Get("local_login_enabled")) == "1"
	minLen, err := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("password_min_length")))
	if err != nil || minLen < 6 || minLen > 72 {
		redirectWithWebMessage(w, r, settingsAuthRoute, "auth_settings_invalid")
		return
	}
	localVal := "false"
	if localEnabled {
		localVal = "true"
	}
	if err := orm.SetConfigParam(ctx, authLocalConfigKey, localVal); err != nil {
		redirectWithWebMessage(w, r, settingsAuthRoute, "auth_settings_failed")
		return
	}
	if err := orm.SetConfigParam(ctx, authPasswordMinLengthKey, strconv.Itoa(minLen)); err != nil {
		redirectWithWebMessage(w, r, settingsAuthRoute, "auth_settings_failed")
		return
	}
	redirectWithWebMessage(w, r, settingsAuthRoute, "auth_settings_saved")
}

func SettingsMailGetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSystemAdmin(w, r, true) {
		return
	}
	ctx := r.Context()
	menuIDStr, ok := resolveSettingsRootMenuID(w, r, ctx)
	if !ok {
		return
	}
	flash, _ := FlashFromQueryMessage(r.URL.Query().Get("msg"))
	renderSettingsMailPage(w, r, loadSettingsMailData(CSRFTokenForRequest(r), flash), menuIDStr)
}

func loadSettingsAuthData(ctx context.Context, csrf string, flash render.FlashMessage) settingsAuthData {
	localRaw := strings.ToLower(strings.TrimSpace(orm.GetConfig(ctx, authLocalConfigKey, "true")))
	minRaw := strings.TrimSpace(orm.GetConfig(ctx, authPasswordMinLengthKey, authPasswordMinLengthDefault))
	minLen, _ := strconv.Atoi(minRaw)
	if minLen < 6 {
		minLen = 8
	}
	return settingsAuthData{
		CSRFToken:         csrf,
		Flash:             flash,
		LocalLoginEnabled: localRaw != "false" && localRaw != "0",
		PasswordMinLength: minLen,
	}
}

func loadSettingsMailData(csrf string, flash render.FlashMessage) settingsMailData {
	cfg := config.AppConfig
	port := cfg.SMTPPort
	if port <= 0 {
		port = 587
	}
	return settingsMailData{
		CSRFToken:  csrf,
		Flash:      flash,
		Configured: mail.Configured(),
		SMTPHost:   strings.TrimSpace(cfg.SMTPHost),
		SMTPPort:   port,
		SMTPFrom:   strings.TrimSpace(cfg.SMTPFrom),
		SMTPUser:   strings.TrimSpace(cfg.SMTPUser),
	}
}

func renderSettingsAuthPage(w http.ResponseWriter, r *http.Request, pageData settingsAuthData, menuIDStr string) {
	ctx := r.Context()
	renderShellPage(w, r, shellPageOpts{
		Route:         settingsAuthRoute,
		InnerTemplate: settingsAuthInnerTemplate,
		InnerData:     pageData,
		MenuIDStr:     menuIDStr,
		Page:          buildSettingsPlatformPageData(ctx, menuIDStr, "Authentication", pageData.Flash),
	})
}

func renderSettingsMailPage(w http.ResponseWriter, r *http.Request, pageData settingsMailData, menuIDStr string) {
	ctx := r.Context()
	renderShellPage(w, r, shellPageOpts{
		Route:         settingsMailRoute,
		InnerTemplate: settingsMailInnerTemplate,
		InnerData:     pageData,
		MenuIDStr:     menuIDStr,
		Page:          buildSettingsPlatformPageData(ctx, menuIDStr, "Mail delivery", pageData.Flash),
	})
}

func buildSettingsPlatformPageData(ctx context.Context, menuIDStr, sectionTitle string, flash render.FlashMessage) render.PageData {
	crumbs := render.BuildSettingsHubBreadcrumbs(ctx)
	crumbs = append(crumbs, render.BreadcrumbItem{Label: sectionTitle})
	pd := render.PageData{
		Title:              sectionTitle,
		SettingsNavActive:  true,
		ActiveMenuID:       menuIDStr,
		BreadcrumbItems:    crumbs,
		ViewStylesheetURLs: []string{settingsHubStylesheetURL},
		ExtraBodyClasses:   settingsHubBodyClass,
	}
	if flash.Body != "" || flash.Title != "" {
		pd.FlashMessages = []render.FlashMessage{flash}
	}
	return pd
}

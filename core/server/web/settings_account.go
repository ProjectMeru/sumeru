package web

import (
	"context"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

const settingsAccountRoute = "/web/settings/account"

func registerSettingsAccountRoutes() {
	registerSession(http.MethodGet, settingsAccountRoute, SettingsAccountGetHandler)
	registerSession(http.MethodPost, settingsAccountRoute, SettingsAccountPostHandler)
	registerSettingsAccountTOTPRoutes()
}

func SettingsAccountGetHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSettingsUser(w, r) {
		return
	}
	ctx := r.Context()
	menuIDStr, ok := resolveSettingsRootMenuID(w, r, ctx)
	if !ok {
		return
	}
	flash, _ := FlashFromQueryMessage(r.URL.Query().Get("msg"))
	actor := orm.SecurityUID(ctx)
	renderSettingsAccountPage(w, r, settingsAccountData{
		CSRFToken:          CSRFTokenForRequest(r),
		Flash:              flash,
		TOTP:               loadSettingsAccountTOTPState(ctx, actor),
		APIKeys:            loadSettingsAccountAPIKeys(ctx, actor),
		MaxKeyDurationDays: orm.MaxAPIKeyDurationDays,
	}, menuIDStr)
}

func SettingsAccountPostHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireSettingsUser(w, r) {
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
	actor := orm.SecurityUID(ctx)
	action := strings.TrimSpace(r.PostForm.Get("action"))
	if handleSettingsAccountTOTP(w, r, ctx, actor, action) {
		return
	}
	if handleSettingsAccountAPIKey(w, r, ctx, actor, action) {
		return
	}
	if action != "" {
		http.Error(w, "unknown action", http.StatusBadRequest)
		return
	}
	pw := strings.TrimSpace(r.PostForm.Get("password_plain"))
	confirm := strings.TrimSpace(r.PostForm.Get("password_plain_confirm"))
	if pw == "" {
		redirectWithWebMessage(w, r, settingsAccountRoute, "password_required")
		return
	}
	if pw != confirm {
		redirectWithWebMessage(w, r, settingsAccountRoute, "password_mismatch")
		return
	}
	if err := orm.SetOwnUserPassword(ctx, actor, actor, pw); err != nil {
		redirectWithWebMessage(w, r, settingsAccountRoute, "password_failed")
		return
	}
	redirectWithWebMessage(w, r, settingsAccountRoute, "password_updated")
}

type settingsAccountData struct {
	CSRFToken          string
	Flash              render.FlashMessage
	TOTP               settingsAccountTOTPState
	APIKeys            []settingsAccountAPIKeyRow
	MaxKeyDurationDays int
}

func renderSettingsAccountPage(w http.ResponseWriter, r *http.Request, pageData settingsAccountData, menuIDStr string) {
	ctx := r.Context()
	renderShellPage(w, r, shellPageOpts{
		Route:         settingsAccountRoute,
		InnerTemplate: settingsAccountInnerTemplate,
		InnerData:     pageData,
		MenuIDStr:     menuIDStr,
		Page:          buildSettingsAccountPageData(ctx, menuIDStr, pageData.Flash),
	})
}

func buildSettingsAccountPageData(ctx context.Context, menuIDStr string, flash render.FlashMessage) render.PageData {
	crumbs := render.BuildSettingsHubBreadcrumbs(ctx)
	crumbs = append(crumbs, render.BreadcrumbItem{Label: "Account security"})
	pd := render.PageData{
		Title:              "Account security",
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

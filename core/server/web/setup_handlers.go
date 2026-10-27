package web

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/errcode"
	"sumeru/core/module"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const setupRestartDelay = 400 * time.Millisecond

type setupInitRequest struct {
	CompanyName string `json:"company_name"`
	Lang        string `json:"lang"`
	AdminName   string `json:"admin_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	SetupToken  string `json:"setup_token"`
	CSRFToken   string `json:"csrf_token"`
}

type setupPageData struct {
	DbName             string
	Stylesheets        []string
	SetupTokenRequired bool
	AppName            string
	Brand              loginBrand
	CSRFToken          string
}

// SetupInitHandler runs database sync, installs base, bootstraps security from the JSON wizard payload, then restarts.
func SetupInitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !RequirePOST(w, r) {
		return
	}

	requestBody, readOK := readBoundedRequestBody(r, maxSetupInitBodyBytes)
	if !readOK {
		http.Error(w, "Could not read body", http.StatusBadRequest)
		return
	}
	if int64(len(requestBody)) > maxSetupInitBodyBytes {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	payload, ok := parseSetupInitRequest(w, requestBody)
	if !ok {
		return
	}
	if !validateSetupCSRF(r, payload.CSRFToken) {
		http.Error(w, "Invalid request", http.StatusForbidden)
		return
	}
	if !allowSetupRequest(w, r, payload.SetupToken) {
		return
	}

	if err := runFirstTimeSetup(ctx, toSetupAdminParams(payload)); err != nil {
		logSetupFailure(ctx, "Setup init failed", err)
		http.Error(w, "Setup failed", http.StatusInternalServerError)
		return
	}

	scheduleSetupRestart()
	_, _ = fmt.Fprintln(w, setupCompleteMessage)
}

// SetupPageHandler renders the setup page from templates/setup.html.
func SetupPageHandler(w http.ResponseWriter, r *http.Request) {
	if !requireSetupEnvironment(w, r) {
		return
	}
	pageData := buildSetupPageData()
	pageData.CSRFToken = setSetupCSRFCookie(w)
	writeSetupPage(w, r.Context(), pageData)
}

func setSetupCSRFCookie(w http.ResponseWriter) string {
	token := newSetupCSRFToken()
	setNamedCookie(w, setupCSRFCookie, token, setupRoute, 600, http.SameSiteStrictMode)
	return token
}

func newSetupCSRFToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic("setup csrf: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(buf)
}

func validateSetupCSRF(r *http.Request, tokenFromBody string) bool {
	cookie, err := r.Cookie(setupCSRFCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	got := strings.TrimSpace(tokenFromBody)
	if got == "" {
		got = strings.TrimSpace(r.Header.Get(csrfHeaderName))
	}
	if got == "" {
		return false
	}
	return hmac.Equal([]byte(got), []byte(cookie.Value))
}

func parseSetupInitRequest(w http.ResponseWriter, requestBody []byte) (setupInitRequest, bool) {
	var payload setupInitRequest
	if err := json.Unmarshal(requestBody, &payload); err != nil {
		http.Error(w, "Expected JSON with company_name, lang, admin_name, email, password", http.StatusBadRequest)
		return setupInitRequest{}, false
	}
	return payload, true
}

func toSetupAdminParams(payload setupInitRequest) orm.SetupAdminParams {
	return orm.SetupAdminParams{
		CompanyName: payload.CompanyName,
		Lang:        payload.Lang,
		FullName:    payload.AdminName,
		Email:       payload.Email,
		Password:    payload.Password,
	}
}

func runFirstTimeSetup(ctx context.Context, adminParams orm.SetupAdminParams) error {
	return orm.WithElevated(context.Background(), "setup.bootstrap", func(securityContext context.Context) error {
		if err := module.RunFirstTimeInstallSync(securityContext); err != nil {
			logSetupFailure(ctx, "First-time install sync failed", err)
			return err
		}
		if err := module.InstallModuleByName(securityContext, baseModuleName); err != nil {
			logSetupFailure(ctx, "Install base module failed", err)
			return fmt.Errorf("install base failed: %w", err)
		}
		if err := orm.EnsureBootstrapSecurityFromSetup(adminParams); err != nil {
			logSetupFailure(ctx, "Security bootstrap failed", err)
			return fmt.Errorf("security bootstrap failed: %w", err)
		}
		return nil
	})
}

func logSetupFailure(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	ev := applog.Event{
		Component: "web",
		Operation: setupOperation,
		Status:    "failure",
		Err:       err,
	}
	if len(fields) > 0 {
		ev.Context = fields[0]
	}
	applog.ErrorCode(ctx, errcode.InternalError, message, ev)
}

func scheduleSetupRestart() {
	go func() {
		time.Sleep(setupRestartDelay)
		applog.InfoMsg(context.Background(), "web", setupOperation, "Self-restarting server after setup", nil)
		if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
			applog.Fatal(context.Background(), "Setup self-restart failed", "err", err)
		}
	}()
}

func buildSetupPageData() setupPageData {
	return setupPageData{
		DbName:             config.AppConfig.DbName,
		Stylesheets:        assets.LoginStylesheetURLs(),
		SetupTokenRequired: strings.TrimSpace(config.AppConfig.SetupToken) != "",
		AppName:            loginBrandPlatformName,
		Brand:              setupLoginBrand(),
	}
}

func writeSetupPage(w http.ResponseWriter, ctx context.Context, pageData setupPageData) {
	dir := config.AppConfig.TemplatesPath
	paths := render.AuthTemplateFiles(dir, setupTemplateFile)
	templateFile, err := template.ParseFiles(paths...)
	if err != nil {
		logSetupFailure(ctx, "Failed to parse setup template", err, map[string]interface{}{"template": setupTemplateFile})
		http.Error(w, "Setup template missing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templateFile.Execute(w, pageData); err != nil {
		logSetupFailure(ctx, "Failed to execute setup template", err)
	}
}

package web

import (
	"bufio"
	"context"
	"errors"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/mail"
	"sumeru/core/orm"
	"sumeru/core/sdk/platformmsg"
	"sumeru/core/server/config"

	"golang.org/x/crypto/bcrypt"
)

// loginPageData is the view model for templates/login.html.
type loginPageData struct {
	Next        string
	Error       string
	Stylesheets []string
	LogoURL     string
}

// loginUser holds the fields needed to verify a password at sign-in.
type loginUser struct {
	ID           int
	PasswordHash string
	Active       bool
}

type loginCredentials struct {
	Login    string
	Password string
	Next     string
}

var (
	loginTemplateOnce sync.Once
	cachedLoginTmpl   *template.Template
	loginTemplateErr  error
)

// APIKeyUserID resolves X-API-Key or Authorization: Bearer credentials to a user id.
func APIKeyUserID(r *http.Request) int {
	if r == nil {
		return 0
	}
	raw := apiKeyFromRequest(r)
	if raw == "" {
		return 0
	}
	return orm.UIDFromAPIKey(r.Context(), raw)
}

// AuthenticatedUserID returns the session user id, or the API key user id when no session exists.
func AuthenticatedUserID(r *http.Request) int {
	if uid := SessionUserID(r); uid > 0 {
		return uid
	}
	return APIKeyUserID(r)
}

// SecurityMiddleware attaches request_id, authenticated uid, and active company to each request.
func SecurityMiddleware(next http.Handler) http.Handler {
	if next == nil {
		next = http.DefaultServeMux
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !enforceRateLimit(w, r) {
			return
		}
		start := time.Now()
		requestID := requestIDFromHeader(r)
		w.Header().Set(requestIDHeader, requestID)

		ctx := enrichRequestContext(r, requestID)
		r = r.WithContext(ctx)

		logHTTPRequestStart(ctx, r)

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		logHTTPRequestEnd(ctx, r, recorder.status, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(statusCode int) {
	recorder.status = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}

func (recorder *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := recorder.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("web: ResponseWriter does not support hijacking")
	}
	return hj.Hijack()
}

func (recorder *statusRecorder) Flush() {
	if f, ok := recorder.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// requireLogin redirects anonymous browser requests to the login page with a safe return URL.
func requireLogin(w http.ResponseWriter, r *http.Request) bool {
	if SessionUserID(r) > 0 {
		return true
	}
	returnTo := SafePathNext(r.URL.RequestURI(), homeRoute)
	http.Redirect(w, r, loginURLWithReturn(returnTo), http.StatusFound)
	return false
}

// LoginGet renders the login form for anonymous users.
func LoginGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	next := strings.TrimSpace(r.URL.Query().Get(nextField))
	if SessionUserID(r) > 0 {
		http.Redirect(w, r, SafePathNext(next, homeRoute), http.StatusFound)
		return
	}

	writeLoginPage(w, r, http.StatusOK, next, "")
}

// LoginPost validates credentials, opens a session, and redirects to the requested page.
func LoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !ParsePostForm(w, r) {
		return
	}

	credentials := parseLoginCredentials(r)
	clientIP := clientIP(r)

	user, authenticated := verifyLoginCredentials(r.Context(), credentials, clientIP)
	if !authenticated {
		writeLoginPage(w, r, http.StatusUnauthorized, credentials.Next, invalidLoginMessage)
		return
	}
	if err := CreateSession(w, user.ID); err != nil {
		WebLogf(r.Context(), loginRoute, "session: %v", err)
		http.Error(w, "Could not start session", http.StatusInternalServerError)
		return
	}

	orm.AppendUserLog(r.Context(), user.ID, clientIP, "success")
	http.Redirect(w, r, credentials.Next, http.StatusSeeOther)
}

// LogoutGet destroys the session cookie and returns the browser to the login page.
func LogoutGet(w http.ResponseWriter, r *http.Request) {
	DestroySession(w, r)
	http.Redirect(w, r, loginRoute, http.StatusFound)
}

// ActionResetPassword accepts a reset request from a system administrator (email delivery not yet wired).
func ActionResetPassword(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	if !requireSystemAdmin(w, r, false) {
		return
	}

	userID := strings.TrimSpace(r.PostFormValue(resetUserIDField))
	loginName := strings.TrimSpace(r.PostFormValue(loginField))
	to := strings.TrimSpace(r.PostFormValue("email"))
	if to == "" && strings.Contains(loginName, "@") {
		to = loginName
	}
	loginURL := loginRoute
	if mail.Configured() && to != "" {
		if err := mail.SendPasswordResetEmail(r.Context(), to, loginName, loginURL); err != nil {
			WebLogf(r.Context(), resetPasswordRoute, "email failed for user id=%s: %v", userID, err)
		} else {
			WebLogf(r.Context(), resetPasswordRoute, "reset email sent for user id=%s login=%q", userID, loginName)
		}
	} else {
		WebLogf(r.Context(), resetPasswordRoute,
			"requested for user id=%s login=%q (configure smtp_host/smtp_from to send email)", userID, loginName)
	}
	redirectWithWebMessage(w, r, r.PostFormValue(nextField), resetPasswordMsg)
}

func loginURLWithReturn(returnTo string) string {
	return loginRoute + "?next=" + url.QueryEscape(returnTo)
}

func parseLoginCredentials(r *http.Request) loginCredentials {
	return loginCredentials{
		Login:    strings.TrimSpace(r.PostFormValue(loginField)),
		Password: r.PostFormValue(passwordField),
		Next:     SafePathNext(r.PostFormValue(nextField), homeRoute),
	}
}

func verifyLoginCredentials(ctx context.Context, credentials loginCredentials, clientIP string) (loginUser, bool) {
	user, err := lookupLoginUser(ctx, credentials.Login)
	if err != nil || !userCanAuthenticate(user) {
		recordFailedLogin(ctx, 0, clientIP, "login="+credentials.Login)
		return loginUser{}, false
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
		recordFailedLogin(ctx, user.ID, clientIP, "bad password")
		return loginUser{}, false
	}
	return user, true
}

func apiKeyFromRequest(r *http.Request) string {
	if key := strings.TrimSpace(r.Header.Get(apiKeyHeader)); key != "" {
		return key
	}
	return bearerToken(r.Header.Get(authHeader))
}

func bearerToken(authHeaderValue string) string {
	authValue := strings.TrimSpace(authHeaderValue)
	if len(authValue) < len(authBearerPrefix) || !strings.EqualFold(authValue[:len(authBearerPrefix)], authBearerPrefix) {
		return ""
	}
	return strings.TrimSpace(authValue[len(authBearerPrefix):])
}

func requestIDFromHeader(r *http.Request) string {
	if requestID := strings.TrimSpace(r.Header.Get(requestIDHeader)); requestID != "" {
		return requestID
	}
	return applog.NewRequestID()
}

func enrichRequestContext(r *http.Request, requestID string) context.Context {
	ctx := applog.ContextWithRequestID(r.Context(), requestID)
	userID := AuthenticatedUserID(r)
	ctx = orm.ContextWithUID(ctx, userID)
	if userID > 0 {
		ctx = orm.ContextWithCompanyID(ctx, orm.ActiveCompanyIDForUser(ctx, userID))
	}
	return ctx
}

func logHTTPRequestStart(ctx context.Context, r *http.Request) {
	applog.Debug(ctx, applog.Event{
		Message:   "HTTP request started",
		Component: "web",
		Operation: "request",
		Status:    "success",
		Context: map[string]interface{}{
			"route":  r.URL.Path,
			"method": r.Method,
		},
	})
}

func logHTTPRequestEnd(ctx context.Context, r *http.Request, statusCode int, duration time.Duration) {
	event := applog.Event{
		Component: "web",
		Operation: "request",
		Duration:  duration,
		Context: map[string]interface{}{
			"route":       r.URL.Path,
			"method":      r.Method,
			"status_code": statusCode,
		},
	}
	if statusCode >= 500 {
		event.Message = "HTTP request failed"
		event.Status = "failure"
		applog.Error(ctx, event)
		return
	}
	event.Message = "HTTP request completed"
	event.Status = "success"
	applog.Debug(ctx, event)
}

func getLoginTemplate() (*template.Template, error) {
	loginTemplateOnce.Do(func() {
		templatePath := filepath.Join(config.AppConfig.TemplatesPath, loginTemplateFile)
		cachedLoginTmpl, loginTemplateErr = template.ParseFiles(templatePath)
	})
	return cachedLoginTmpl, loginTemplateErr
}

func newLoginPageData(next, errorMessage string) loginPageData {
	return loginPageData{
		Next:        next,
		Error:       errorMessage,
		Stylesheets: assets.LoginStylesheetURLs(),
		LogoURL:     render.ShellLogoURL(),
	}
}

func writeLoginPage(w http.ResponseWriter, r *http.Request, statusCode int, next, errorMessage string) {
	tmpl, err := getLoginTemplate()
	if err != nil {
		if statusCode == http.StatusOK {
			WebLogf(r.Context(), loginRoute, "%s: login template: %v", platformmsg.MsgHTTPTemplateError, err)
			http.Error(w, "Login page unavailable", http.StatusInternalServerError)
			return
		}
		http.Error(w, errorMessage, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	_ = tmpl.Execute(w, newLoginPageData(next, errorMessage))
}

func lookupLoginUser(ctx context.Context, loginName string) (loginUser, error) {
	userTable := orm.MustQuotedTableName(coreUserModel)
	var user loginUser
	err := orm.DB.QueryRowContext(ctx,
		`SELECT id, COALESCE(password, ''), active FROM `+userTable+` WHERE LOWER(TRIM(login)) = LOWER(TRIM($1)) LIMIT 1`,
		loginName,
	).Scan(&user.ID, &user.PasswordHash, &user.Active)
	return user, err
}

func userCanAuthenticate(user loginUser) bool {
	return user.Active && strings.TrimSpace(user.PasswordHash) != ""
}

func recordFailedLogin(ctx context.Context, userID int, clientIP, auditNote string) {
	orm.AppendUserLog(ctx, userID, clientIP, "failure")
	orm.AppendAudit(ctx, "login_fail", coreUserModel, int64(userID), nil, nil, auditNote)
}

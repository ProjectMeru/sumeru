package web

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/errcode"
	"sumeru/core/mail"
	"sumeru/core/orm"
	"sumeru/core/server/auth"
	"sumeru/core/server/config"
)

const loginLockoutWindow = 15 * time.Minute

type loginPageData struct {
	Next              string
	Error             string
	CSRFToken         string
	Stylesheets       []string
	Brand             loginBrand
	AuthProviders     []loginAuthProvider
	LocalLoginEnabled bool
	AppName           string
	InfoTitle         string
	InfoBody          string
	Year              int
}

type loginCredentials struct {
	Login    string
	Password string
	Next     string
}

type loginFinishOpts struct {
	UserID   int
	Next     string
	LoginKey string
	ClientIP string
	LogRoute string
}

// ponytail: in-memory lockout map; single-process only; upgrade to DB/redis for multi-instance.
var (
	loginLockoutMu    sync.Mutex
	loginFailures     = map[string][]time.Time{}
	authTemplateMu    sync.Mutex
	authTemplateCache = map[string]*template.Template{}
	authTemplateErrs  = map[string]error{}
)

func setLoginCSRFCookie(w http.ResponseWriter) string {
	token := newLoginCSRFToken()
	setNamedCookie(w, loginCSRFCookie, token, loginRoute, 600, http.SameSiteStrictMode)
	return token
}

func newLoginCSRFToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic("login csrf: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(buf)
}

func validateLoginCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(loginCSRFCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	got := strings.TrimSpace(r.PostFormValue(csrfFormField))
	if got == "" {
		return false
	}
	return hmac.Equal([]byte(got), []byte(cookie.Value))
}

func clearLoginCSRFCookie(w http.ResponseWriter) {
	clearNamedCookie(w, loginCSRFCookie, loginRoute, http.SameSiteStrictMode)
}

func setLoginNextCookie(w http.ResponseWriter, returnTo string) {
	returnTo = SafePathNext(returnTo, homeRoute)
	setNamedCookie(w, loginNextCookie, returnTo, loginRoute, 600, http.SameSiteLaxMode)
}

func loginNextFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(loginNextCookie)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return SafePathNext(cookie.Value, homeRoute)
}

func resolveLoginNext(r *http.Request) string {
	var next string
	if n := loginNextFromRequest(r); n != "" {
		next = n
	} else if q := strings.TrimSpace(r.URL.Query().Get(nextField)); q != "" {
		next = SafePathNext(q, homeRoute)
	} else {
		next = homeRoute
	}
	if uid := SessionUserID(r); uid > 0 {
		return postLoginDestination(r.Context(), uid, next)
	}
	return next
}

func clearLoginNextCookie(w http.ResponseWriter) {
	clearNamedCookie(w, loginNextCookie, loginRoute, http.SameSiteLaxMode)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, returnTo string) {
	setLoginNextCookie(w, returnTo)
	http.Redirect(w, r, loginRoute, http.StatusFound)
}

func normalizeLoginKey(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

func loginLocked(login string) bool {
	key := normalizeLoginKey(login)
	if key == "" {
		return false
	}
	loginLockoutMu.Lock()
	defer loginLockoutMu.Unlock()
	cutoff := time.Now().Add(-loginLockoutWindow)
	attempts := pruneAttemptsAfter(loginFailures[key], cutoff)
	loginFailures[key] = attempts
	return len(attempts) >= loginLockoutMaxFailures
}

func recordLoginFailure(login string) {
	key := normalizeLoginKey(login)
	if key == "" {
		return
	}
	loginLockoutMu.Lock()
	defer loginLockoutMu.Unlock()
	cutoff := time.Now().Add(-loginLockoutWindow)
	loginFailures[key] = append(pruneAttemptsAfter(loginFailures[key], cutoff), time.Now())
}

func clearLoginFailures(login string) {
	key := normalizeLoginKey(login)
	if key == "" {
		return
	}
	loginLockoutMu.Lock()
	delete(loginFailures, key)
	loginLockoutMu.Unlock()
}

func resetLoginLockoutState() {
	loginLockoutMu.Lock()
	loginFailures = map[string][]time.Time{}
	loginLockoutMu.Unlock()
}

func comparePasswordConstantTime(storedHash, plain string) bool {
	return auth.ComparePasswordHash(storedHash, plain)
}

func parsePositiveInt(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func resetAuthTemplateCache() {
	authTemplateMu.Lock()
	defer authTemplateMu.Unlock()
	authTemplateCache = map[string]*template.Template{}
	authTemplateErrs = map[string]error{}
}

func getAuthTemplate(filename string) (*template.Template, error) {
	authTemplateMu.Lock()
	defer authTemplateMu.Unlock()
	if _, loaded := authTemplateErrs[filename]; loaded {
		return authTemplateCache[filename], authTemplateErrs[filename]
	}
	dir := config.AppConfig.TemplatesPath
	paths := render.AuthTemplateFiles(dir, filename)
	tmpl, err := template.ParseFiles(paths...)
	authTemplateCache[filename] = tmpl
	authTemplateErrs[filename] = err
	return tmpl, err
}

func buildLoginPageData(r *http.Request, next, errorMessage, csrfToken string) loginPageData {
	ctx := r.Context()
	data := loginPageData{
		Next:              next,
		Error:             errorMessage,
		CSRFToken:         csrfToken,
		Stylesheets:       assets.LoginStylesheetURLs(),
		Brand:             resolveLoginBrand(ctx),
		AuthProviders:     listEnabledAuthProviders(ctx),
		LocalLoginEnabled: authLocalEnabled(ctx),
		AppName:           loginBrandPlatformName,
		Year:              time.Now().Year(),
	}
	if flash, ok := FlashFromQueryMessage(r.URL.Query().Get(flashMessageParam)); ok {
		if errorMessage == "" && flash.Kind == "error" {
			data.Error = flash.Body
			if flash.Title != "" {
				data.Error = flash.Title + ": " + flash.Body
			}
		} else if flash.Body != "" {
			data.InfoTitle = flash.Title
			data.InfoBody = flash.Body
		}
	}
	return data
}

func writeAuthFormPage(w http.ResponseWriter, r *http.Request, templateFile, logRoute string, statusCode int, next, errorMessage, csrfToken string, adjust func(*loginPageData)) {
	tmpl, err := getAuthTemplate(templateFile)
	if err != nil {
		if statusCode == http.StatusOK {
			WebLogEvent(r.Context(), WebLogInput{
				Route:     logRoute,
				Message:   "auth form template unavailable",
				Code:      errcode.InternalError,
				Operation: "login_template",
				Status:    logStatusFailure,
				Err:       err,
			})
			http.Error(w, "Page unavailable", http.StatusInternalServerError)
			return
		}
		http.Error(w, errorMessage, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	data := buildLoginPageData(r, next, errorMessage, csrfToken)
	if adjust != nil {
		adjust(&data)
	}
	_ = tmpl.Execute(w, data)
}

func showAuthFormError(w http.ResponseWriter, r *http.Request, templateFile, logRoute string, statusCode int, next, errorMessage string) {
	csrfToken := setLoginCSRFCookie(w)
	writeAuthFormPage(w, r, templateFile, logRoute, statusCode, next, errorMessage, csrfToken, nil)
}

func registerTOTPRoutes() {
	registerPublic(http.MethodGet, totpLoginRoute, TOTPLoginGet)
	registerPublic(http.MethodPost, totpLoginRoute, TOTPLoginPost)
}

func TOTPLoginGet(w http.ResponseWriter, r *http.Request) {
	if pendingMFAUserID(r) <= 0 {
		http.Redirect(w, r, loginRoute, http.StatusFound)
		return
	}
	csrfToken := setLoginCSRFCookie(w)
	writeAuthFormPage(w, r, totpLoginTemplateFile, totpLoginRoute, http.StatusOK, resolveLoginNext(r), "", csrfToken, func(data *loginPageData) {
		data.Brand = totpLoginBrand(data.Brand)
	})
}

func TOTPLoginPost(w http.ResponseWriter, r *http.Request) {
	if !ParsePostForm(w, r) {
		return
	}
	uid := pendingMFAUserID(r)
	if 0 >= uid {
		http.Error(w, "session expired", http.StatusUnauthorized)
		return
	}
	next := SafePathNext(r.PostFormValue(nextField), homeRoute)
	if !validateLoginCSRF(r) {
		showAuthFormError(w, r, totpLoginTemplateFile, totpLoginRoute, http.StatusForbidden, next, "Invalid form")
		return
	}
	code := strings.TrimSpace(r.PostFormValue("totp_code"))
	secret, _ := userTOTPFields(r.Context(), uid)
	if !auth.ValidateTOTP(secret, code) {
		showAuthFormError(w, r, totpLoginTemplateFile, totpLoginRoute, http.StatusUnauthorized, next, "Invalid authentication code")
		return
	}
	if strings.TrimSpace(r.PostFormValue("trust_device")) == "1" {
		setTrustedDeviceCookie(w, uid)
	}
	finishTOTPSession(w, r, uid, next)
}

func userTOTPFields(ctx context.Context, uid int) (secret string, enabled bool) {
	return orm.UserTOTPCredentialsForLogin(ctx, uid)
}

func needsTOTP(r *http.Request, uid int) bool {
	if trustedDeviceValid(r, uid) {
		return false
	}
	_, enabled := userTOTPFields(r.Context(), uid)
	return enabled
}

func trustedDeviceValid(r *http.Request, expectUID int) bool {
	c, err := r.Cookie(trustedDeviceCookie)
	if err != nil || strings.TrimSpace(c.Value) == "" {
		return false
	}
	cookieUID, ok := parseSignedUIDCookie(c.Value)
	return ok && cookieUID == expectUID
}

func setPendingMFACookie(w http.ResponseWriter, uid int) {
	setNamedCookie(w, pendingMFACookie, mintSignedUIDCookieValue(uid, 5*time.Minute), "/", 300, sessionSameSite())
}

func pendingMFAUserID(r *http.Request) int {
	c, err := r.Cookie(pendingMFACookie)
	if err != nil {
		return 0
	}
	uid, ok := parseSignedUIDCookie(c.Value)
	if !ok {
		return 0
	}
	return uid
}

func clearPendingMFACookie(w http.ResponseWriter) {
	clearNamedCookie(w, pendingMFACookie, "/", sessionSameSite())
}

func setTrustedDeviceCookie(w http.ResponseWriter, uid int) {
	setNamedCookie(w, trustedDeviceCookie, mintSignedUIDCookieValue(uid, 30*24*time.Hour), "/", 30*24*3600, sessionSameSite())
}

func completePasswordOrOAuthLogin(w http.ResponseWriter, r *http.Request, opts loginFinishOpts) {
	next := SafePathNext(opts.Next, homeRoute)
	if needsTOTP(r, opts.UserID) {
		setPendingMFACookie(w, opts.UserID)
		if opts.LoginKey != "" {
			clearLoginFailures(opts.LoginKey)
		}
		clearLoginCSRFCookie(w)
		http.Redirect(w, r, totpLoginRoute+"?next="+url.QueryEscape(next), http.StatusSeeOther)
		return
	}
	if opts.LoginKey != "" {
		clearLoginFailures(opts.LoginKey)
	}
	if opts.ClientIP != "" {
		orm.AppendUserLog(r.Context(), opts.UserID, opts.ClientIP, "success")
	}
	establishSessionAndRedirect(w, r, opts.UserID, next, opts.LogRoute)
}

func finishTOTPSession(w http.ResponseWriter, r *http.Request, uid int, nextFromForm string) {
	clearPendingMFACookie(w)
	establishSessionAndRedirect(w, r, uid, SafePathNext(nextFromForm, homeRoute), totpLoginRoute)
}

func establishSessionAndRedirect(w http.ResponseWriter, r *http.Request, userID int, next, logRoute string) {
	if err := CreateSession(w, r.Context(), userID); err != nil {
		WebLogEvent(r.Context(), WebLogInput{
			Route:     logRoute,
			Message:   "Could not start session",
			Code:      errcode.InternalError,
			Operation: "session_create",
			Status:    logStatusFailure,
			Err:       err,
		})
		http.Error(w, "Could not start session", http.StatusInternalServerError)
		return
	}
	clearLoginCSRFCookie(w)
	clearLoginNextCookie(w)
	dest := postLoginDestination(r.Context(), userID, next)
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func LoginGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if q := strings.TrimSpace(r.URL.Query().Get(nextField)); q != "" {
		setLoginNextCookie(w, q)
		http.Redirect(w, r, loginRoute, http.StatusFound)
		return
	}
	if SessionUserID(r) > 0 {
		http.Redirect(w, r, resolveLoginNext(r), http.StatusFound)
		return
	}

	csrfToken := setLoginCSRFCookie(w)
	writeAuthFormPage(w, r, loginTemplateFile, loginRoute, http.StatusOK, resolveLoginNext(r), "", csrfToken, nil)
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !ParsePostForm(w, r) {
		return
	}
	if !authLocalEnabled(r.Context()) {
		showAuthFormError(w, r, loginTemplateFile, loginRoute, http.StatusForbidden,
			SafePathNext(r.PostFormValue(nextField), homeRoute), "Password sign-in is disabled. Use single sign-on.")
		return
	}
	if !validateLoginCSRF(r) {
		showAuthFormError(w, r, loginTemplateFile, loginRoute, http.StatusForbidden,
			SafePathNext(r.PostFormValue(nextField), homeRoute), "Invalid or expired login form")
		return
	}

	credentials := parseLoginCredentials(r)
	if loginLocked(credentials.Login) {
		showAuthFormError(w, r, loginTemplateFile, loginRoute, http.StatusUnauthorized, credentials.Next, invalidLoginMessage)
		return
	}
	clientIP := clientIP(r)

	userID, ok := verifyLoginCredentials(r.Context(), credentials, clientIP)
	if !ok {
		applog.WarnCode(r.Context(), errcode.InvalidCredentials, "Invalid login or password", applog.Event{
			Component: "web",
			Operation: "login",
			Status:    "failure",
			Context: map[string]interface{}{
				"route": loginRoute,
				"ip":    clientIP,
			},
		})
		showAuthFormError(w, r, loginTemplateFile, loginRoute, http.StatusUnauthorized, credentials.Next, invalidLoginMessage)
		return
	}
	completePasswordOrOAuthLogin(w, r, loginFinishOpts{
		UserID:   userID,
		Next:     credentials.Next,
		LoginKey: credentials.Login,
		ClientIP: clientIP,
		LogRoute: loginRoute,
	})
}

func LogoutGet(w http.ResponseWriter, r *http.Request) {
	// GET must not destroy the session (prefetch, keepalive, bookmarks). Use LogoutPost.
	http.Redirect(w, r, loginRoute, http.StatusFound)
}

func LogoutPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !ParsePostForm(w, r) || !validateSessionCSRF(w, r) {
		return
	}
	DestroySession(w, r)
	clearLoginNextCookie(w)
	http.Redirect(w, r, loginRoute, http.StatusSeeOther)
}

func parseLoginCredentials(r *http.Request) loginCredentials {
	next := SafePathNext(r.PostFormValue(nextField), "")
	if next == "" {
		next = loginNextFromRequest(r)
	}
	if next == "" {
		next = homeRoute
	}
	return loginCredentials{
		Login:    strings.TrimSpace(r.PostFormValue(loginField)),
		Password: r.PostFormValue(passwordField),
		Next:     next,
	}
}

func verifyLoginCredentials(ctx context.Context, credentials loginCredentials, clientIP string) (int, bool) {
	userID, _, _, err := orm.AuthenticateLocalUser(ctx, credentials.Login, credentials.Password)
	if err != nil {
		recordFailedLogin(ctx, 0, clientIP, "login="+credentials.Login)
		recordLoginFailure(credentials.Login)
		return 0, false
	}
	return userID, true
}

func recordFailedLogin(ctx context.Context, userID int, clientIP, auditNote string) {
	orm.AppendUserLog(ctx, userID, clientIP, "failure")
	orm.AppendAudit(ctx, "login_fail", coreUserModel, int64(userID), nil, nil, auditNote)
}

func ActionResetPassword(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			http.NotFound(w, r)
			return
		}
		writePasswordResetForm(w, token, "")
		return
	case http.MethodPost:
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !ParsePostForm(w, r) {
		return
	}
	if token := strings.TrimSpace(r.PostFormValue("token")); token != "" {
		if !validateLoginCSRF(r) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		err := orm.ConsumePasswordReset(r.Context(), token, r.PostFormValue("password"), r.PostFormValue("confirm"))
		if err != nil {
			writePasswordResetForm(w, token, err.Error())
			return
		}
		http.Redirect(w, r, loginRoute, http.StatusSeeOther)
		return
	}
	if !requireLogin(w, r) || !validateSessionCSRF(w, r) || !requireSystemAdmin(w, r, false) {
		return
	}

	userID := strings.TrimSpace(r.PostFormValue(resetUserIDField))
	loginName := strings.TrimSpace(r.PostFormValue(loginField))
	to := strings.TrimSpace(r.PostFormValue("email"))
	if to == "" && strings.Contains(loginName, "@") {
		to = loginName
	}
	uid, _ := strconv.Atoi(userID)
	if mail.Configured() && to != "" && uid > 0 {
		raw, err := orm.IssuePasswordReset(r.Context(), uid)
		if err != nil {
			WebLogEvent(r.Context(), WebLogInput{
				Route: resetPasswordRoute, Message: "password reset token failed",
				Code: errcode.InternalError, Operation: "password_reset", Status: logStatusFailure, Err: err,
				ContextFields: map[string]interface{}{"user_id": userID},
			})
		} else if err := mail.SendPasswordResetEmail(r.Context(), to, loginName, orm.PasswordResetURL(raw)); err != nil {
			WebLogEvent(r.Context(), WebLogInput{
				Route: resetPasswordRoute, Message: "password reset email failed",
				Code: errcode.InternalError, Operation: "password_reset", Status: logStatusFailure, Err: err,
				ContextFields: map[string]interface{}{"user_id": userID},
			})
		} else {
			WebLogf(r.Context(), resetPasswordRoute, "password reset email sent for user id=%s login=%q", userID, loginName)
		}
	} else {
		WebLogf(r.Context(), resetPasswordRoute,
			"password reset notify for user id=%s login=%q (configure smtp_host/smtp_from to send email; this does not reset passwords)", userID, loginName)
	}
	redirectWithWebMessage(w, r, r.PostFormValue(nextField), resetPasswordMsg)
}

func writePasswordResetForm(w http.ResponseWriter, token, errMsg string) {
	csrf := setLoginCSRFCookie(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := template.Must(template.New("reset").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Choose a password</title></head>
<body class="sum-body">
<form method="post" action="{{.Action}}">
<input type="hidden" name="csrf_token" value="{{.CSRF}}"/>
<input type="hidden" name="token" value="{{.Token}}"/>
{{if .Error}}<p>{{.Error}}</p>{{end}}
<label>New password <input type="password" name="password" required autocomplete="new-password"/></label>
<label>Confirm <input type="password" name="confirm" required autocomplete="new-password"/></label>
<button type="submit">Save password</button>
</form>
</body></html>`))
	_ = page.Execute(w, map[string]string{
		"Action": resetPasswordRoute,
		"CSRF":   csrf,
		"Token":  token,
		"Error":  errMsg,
	})
}

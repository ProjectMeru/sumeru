package web

import (
	"context"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"sumeru/core/applog"
	"sumeru/core/engine/assets"
	"sumeru/core/engine/render"
	"sumeru/core/errcode"
	"sumeru/core/mail"
	"sumeru/core/orm"
	"sumeru/core/server/config"

	"golang.org/x/crypto/bcrypt"
)

type loginPageData struct {
	Next        string
	Error       string
	Stylesheets []string
	LogoURL     string
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
		writeLoginPage(w, r, http.StatusUnauthorized, credentials.Next, invalidLoginMessage)
		return
	}
	if err := CreateSession(w, userID); err != nil {
		WebLogEvent(r.Context(), WebLogInput{
			Route:     loginRoute,
			Message:   "Could not start session",
			Code:      errcode.InternalError,
			Operation: "session_create",
			Status:    logStatusFailure,
			Err:       err,
		})
		http.Error(w, "Could not start session", http.StatusInternalServerError)
		return
	}

	orm.AppendUserLog(r.Context(), userID, clientIP, "success")
	http.Redirect(w, r, credentials.Next, http.StatusSeeOther)
}

func LogoutGet(w http.ResponseWriter, r *http.Request) {
	DestroySession(w, r)
	http.Redirect(w, r, loginRoute, http.StatusFound)
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

func verifyLoginCredentials(ctx context.Context, credentials loginCredentials, clientIP string) (int, bool) {
	var userID int
	var passwordHash string
	var active bool
	userTbl := orm.MustQuotedTableName(coreUserModel)
	err := orm.DB.QueryRowContext(ctx,
		`SELECT id, COALESCE(password, ''), active FROM `+userTbl+` WHERE LOWER(TRIM(login)) = LOWER(TRIM($1)) LIMIT 1`,
		credentials.Login,
	).Scan(&userID, &passwordHash, &active)
	if err != nil || !active || strings.TrimSpace(passwordHash) == "" {
		recordFailedLogin(ctx, 0, clientIP, "login="+credentials.Login)
		return 0, false
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(credentials.Password)) != nil {
		recordFailedLogin(ctx, userID, clientIP, "bad password")
		return 0, false
	}
	return userID, true
}

func recordFailedLogin(ctx context.Context, userID int, clientIP, auditNote string) {
	orm.AppendUserLog(ctx, userID, clientIP, "failure")
	orm.AppendAudit(ctx, "login_fail", coreUserModel, int64(userID), nil, nil, auditNote)
}

func getLoginTemplate() (*template.Template, error) {
	loginTemplateOnce.Do(func() {
		templatePath := filepath.Join(config.AppConfig.TemplatesPath, loginTemplateFile)
		cachedLoginTmpl, loginTemplateErr = template.ParseFiles(templatePath)
	})
	return cachedLoginTmpl, loginTemplateErr
}

func writeLoginPage(w http.ResponseWriter, r *http.Request, statusCode int, next, errorMessage string) {
	tmpl, err := getLoginTemplate()
	if err != nil {
		if statusCode == http.StatusOK {
			WebLogEvent(r.Context(), WebLogInput{
				Route:     loginRoute,
				Message:   "login template unavailable",
				Code:      errcode.InternalError,
				Operation: "login_template",
				Status:    logStatusFailure,
				Err:       err,
			})
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
	_ = tmpl.Execute(w, loginPageData{
		Next:        next,
		Error:       errorMessage,
		Stylesheets: assets.LoginStylesheetURLs(),
		LogoURL:     render.ShellLogoURL(),
	})
}

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
			WebLogEvent(r.Context(), WebLogInput{
				Route:     resetPasswordRoute,
				Message:   "login-link email failed",
				Code:      errcode.InternalError,
				Operation: "login_link_email",
				Status:    logStatusFailure,
				Err:       err,
				ContextFields: map[string]interface{}{
					"user_id": userID,
				},
			})
		} else {
			WebLogf(r.Context(), resetPasswordRoute, "login-link email sent for user id=%s login=%q", userID, loginName)
		}
	} else {
		WebLogf(r.Context(), resetPasswordRoute,
			"login-link notify for user id=%s login=%q (configure smtp_host/smtp_from to send email; this does not reset passwords)", userID, loginName)
	}
	redirectWithWebMessage(w, r, r.PostFormValue(nextField), resetPasswordMsg)
}

package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/errcode"
	"sumeru/core/orm"
	"sumeru/core/server/api"
	"sumeru/core/server/config"
)

type authLoginRequest struct {
	Context  map[string]interface{} `json:"context"`
	Login    string                 `json:"login"`
	Password string                 `json:"password"`
}

func handleRPCAuthLogin(w http.ResponseWriter, r *http.Request, requestBody []byte, companyID int64) {
	if !authLocalEnabled(r.Context()) {
		api.WriteResponse(w, http.StatusForbidden, api.AuthFail(http.StatusForbidden, "AUTH_DISABLED", "AuthenticationError", "Password sign-in is disabled", ""))
		return
	}
	var in authLoginRequest
	if err := json.Unmarshal(requestBody, &in); err != nil {
		api.WriteResponse(w, http.StatusBadRequest, api.Fail(api.CodeInvalidJSON, "invalid json", nil))
		return
	}
	login := strings.TrimSpace(in.Login)
	if login == "" || in.Password == "" {
		api.WriteResponse(w, http.StatusBadRequest, api.Fail(api.CodeValidationError, "login and password are required", nil))
		return
	}
	clientIP := clientIP(r)
	if loginLocked(login) {
		writeAuthLoginFailure(w, r, login, clientIP, 0)
		return
	}
	uid, username, name, err := orm.AuthenticateLocalUser(r.Context(), login, in.Password)
	if err != nil {
		writeAuthLoginFailure(w, r, login, clientIP, 0)
		return
	}
	if !orm.UserAllowedCompany(r.Context(), uid, companyID) {
		api.WriteResponse(w, http.StatusForbidden, api.AuthFail(http.StatusForbidden, "TENANT_DENIED", "AuthenticationError", "Authentication failed", authDebugMessage("tenant not allowed")))
		return
	}
	if needsTOTP(r, uid) {
		api.WriteResponse(w, http.StatusForbidden, api.AuthFail(http.StatusForbidden, "MFA_REQUIRED", "AuthenticationError", "Multi-factor authentication required", authDebugMessage("totp required")))
		return
	}
	clearLoginFailures(login)
	if clientIP != "" {
		orm.AppendUserLog(r.Context(), uid, clientIP, "success")
	}
	sessionID, err := orm.CreateUserSession(r.Context(), uid, orm.SessionCreateOpts{StampBrowserSeen: sessionBrowserPresenceEnabled()})
	if err != nil {
		api.WriteResponse(w, http.StatusInternalServerError, api.Fail(api.CodeInternalError, "Could not start session", nil))
		return
	}
	setSessionCookie(w, sessionID, false)
	rawKey, err := orm.IssueAPIKeyForLogin(r.Context(), uid)
	if err != nil {
		api.WriteResponse(w, http.StatusInternalServerError, api.Fail(api.CodeInternalError, "Could not issue api key", nil))
		return
	}
	data := map[string]interface{}{
		"uid":        uid,
		"username":   username,
		"name":       name,
		"session_id": sessionID,
		"api_key":    rawKey,
	}
	resp := api.SuccessWithMessage(data, "User authenticated successfully")
	api.WriteResponse(w, http.StatusOK, resp)
}

func writeAuthLoginFailure(w http.ResponseWriter, r *http.Request, login, clientIP string, userID int) {
	recordFailedLogin(r.Context(), userID, clientIP, "login="+login)
	recordLoginFailure(login)
	applog.WarnCode(r.Context(), errcode.InvalidCredentials, "Invalid login or password", applog.Event{
		Component: "rpc",
		Operation: "auth.login",
		Status:    "failure",
	})
	debug := authDebugMessage("Authentication failed: invalid username or password.")
	api.WriteResponse(w, http.StatusUnauthorized, api.AuthFail(http.StatusUnauthorized, "INVALID_CREDENTIALS", "AuthenticationError", "Authentication failed", debug))
}

func authDebugMessage(msg string) string {
	if config.AppConfig.DevMode {
		return msg
	}
	if msg == "" {
		return "Authentication failed."
	}
	return "Authentication failed."
}

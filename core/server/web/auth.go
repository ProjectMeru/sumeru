package web

import (
	"bufio"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/errcode"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const sessionCookieName = "sumeru_session"
const sessionDuration = 24 * time.Hour
const sessionSlidingTTL = 8 * time.Hour

var testSessionUserIDOverride int

type ctxKeySessionState struct{}

type sessionState struct {
	userID      int
	fromCookie  bool
	clearCookie bool
}

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
		setSecurityHeaders(w, r)

		session := resolveSession(r)
		if session.clearCookie {
			ClearSessionCookie(w)
		}
		ctx := enrichRequestContext(r, requestID, session)
		r = r.WithContext(ctx)

		logHTTPRequestStart(ctx, r)

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		logHTTPRequestEnd(ctx, r, recorder.status, time.Since(start))
	})
}

func setSecurityHeaders(w http.ResponseWriter, _ *http.Request) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("X-Frame-Options", "SAMEORIGIN")
	h.Set("Content-Security-Policy", "frame-ancestors 'self'; base-uri 'self'; object-src 'none'")
	if !config.AppConfig.DevMode {
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
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

func requestIDFromHeader(r *http.Request) string {
	if requestID := strings.TrimSpace(r.Header.Get(requestIDHeader)); requestID != "" {
		return requestID
	}
	return applog.NewRequestID()
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
		event.Code = errcode.InternalError
		applog.ErrorCode(ctx, event.Code, event.Message, event)
		return
	}
	event.Message = "HTTP request completed"
	event.Status = "success"
	applog.Debug(ctx, event)
}

func buildSessionCookie(value string, deleteCookie bool) *http.Cookie {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !config.AppConfig.DevMode,
	}
	if deleteCookie {
		cookie.MaxAge = -1
	}
	return cookie
}

func withSessionState(ctx context.Context, state sessionState) context.Context {
	return context.WithValue(ctx, ctxKeySessionState{}, state)
}

func sessionStateFrom(ctx context.Context) (sessionState, bool) {
	state, ok := ctx.Value(ctxKeySessionState{}).(sessionState)
	return state, ok
}

func resolveSession(r *http.Request) sessionState {
	if orm.DB == nil {
		return sessionState{}
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return sessionState{}
	}
	sid := cookie.Value
	sessionTbl := orm.MustQuotedTableName("sys.session")
	userTbl := orm.MustQuotedTableName("core.user")

	var userID int
	var active bool
	err = orm.DB.QueryRowContext(r.Context(),
		`SELECT s.user_id, u.active FROM `+sessionTbl+` s
		 JOIN `+userTbl+` u ON u.id = s.user_id
		 WHERE s.sid = $1 AND s.expires_at > NOW()`,
		sid,
	).Scan(&userID, &active)
	if err != nil {
		if err != sql.ErrNoRows {
			applog.WarnCode(r.Context(), errcode.InternalError, "session lookup failed", applog.Event{
				Component: "web",
				Operation: "session_resolve",
				Status:    "partial",
				Err:       err,
			})
		}
		deleteSession(r.Context(), sid)
		return sessionState{clearCookie: true}
	}
	if !active || userID <= 0 {
		deleteSession(r.Context(), sid)
		applog.WarnCode(r.Context(), errcode.AccessDenied, "session revoked for inactive user", applog.Event{
			Component: "web",
			Operation: "session_revoked_inactive",
			Status:    "success",
			Context: map[string]interface{}{
				"user_id": userID,
			},
		})
		orm.AppendAudit(r.Context(), "session_revoked_inactive", "sys.session", 0, nil, nil, fmt.Sprintf("user_id=%d", userID))
		return sessionState{clearCookie: true}
	}
	if _, err := orm.DB.ExecContext(r.Context(),
		`UPDATE `+sessionTbl+` SET expires_at = $1 WHERE sid = $2 AND expires_at > NOW()`,
		time.Now().UTC().Add(sessionSlidingTTL),
		sid,
	); err != nil {
		applog.WarnCode(r.Context(), errcode.InternalError, "sliding session expiry update failed", applog.Event{
			Component: "web",
			Operation: "session_slide",
			Status:    "partial",
			Err:       err,
		})
	}
	return sessionState{userID: userID, fromCookie: true}
}

func deleteSession(ctx context.Context, sid string) {
	if orm.DB == nil || sid == "" {
		return
	}
	sessionTbl := orm.MustQuotedTableName("sys.session")
	if _, err := orm.DB.ExecContext(ctx, `DELETE FROM `+sessionTbl+` WHERE sid = $1`, sid); err != nil {
		applog.WarnCode(ctx, errcode.InternalError, "session delete failed", applog.Event{
			Component: "web",
			Operation: "session_destroy",
			Status:    "partial",
			Err:       err,
		})
	}
}

func CreateSession(w http.ResponseWriter, userID int) error {
	if orm.DB == nil {
		return fmt.Errorf("no database")
	}
	sessionBytes := make([]byte, 24)
	if _, err := rand.Read(sessionBytes); err != nil {
		return err
	}
	sessionID := hex.EncodeToString(sessionBytes)
	expiresAt := time.Now().UTC().Add(sessionDuration)
	sessionTbl := orm.MustQuotedTableName("sys.session")
	if _, err := orm.DB.Exec(`INSERT INTO `+sessionTbl+` (sid, user_id, expires_at) VALUES ($1, $2, $3)`, sessionID, userID, expiresAt); err != nil {
		return err
	}
	http.SetCookie(w, buildSessionCookie(sessionID, false))
	return nil
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, buildSessionCookie("", true))
}

func sessionForRequest(r *http.Request) sessionState {
	if testSessionUserIDOverride > 0 {
		return sessionState{userID: testSessionUserIDOverride, fromCookie: true}
	}
	if state, ok := sessionStateFrom(r.Context()); ok {
		return state
	}
	return resolveSession(r)
}

func SessionUserID(r *http.Request) int {
	return sessionForRequest(r).userID
}

func AuthViaSession(r *http.Request) bool {
	return sessionForRequest(r).fromCookie
}

func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		deleteSession(r.Context(), cookie.Value)
	}
	ClearSessionCookie(w)
}

func APIKeyUserID(r *http.Request) int {
	raw := apiKeyFromRequest(r)
	if raw == "" {
		return 0
	}
	return orm.UIDFromAPIKey(r.Context(), raw)
}

func AuthenticatedUserID(r *http.Request) int {
	if uid := SessionUserID(r); uid > 0 {
		return uid
	}
	return APIKeyUserID(r)
}

func requireLogin(w http.ResponseWriter, r *http.Request) bool {
	if SessionUserID(r) > 0 {
		return true
	}
	returnTo := SafePathNext(r.URL.RequestURI(), homeRoute)
	http.Redirect(w, r, loginURLWithReturn(returnTo), http.StatusFound)
	return false
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

func enrichRequestContext(r *http.Request, requestID string, session sessionState) context.Context {
	ctx := applog.ContextWithRequestID(r.Context(), requestID)
	ctx = withSessionState(ctx, session)
	userID := session.userID
	if userID <= 0 {
		userID = APIKeyUserID(r)
	}
	ctx = orm.ContextWithUID(ctx, userID)
	if userID > 0 {
		ctx = orm.ContextWithCompanyID(ctx, orm.ActiveCompanyIDForUser(ctx, userID))
	}
	return ctx
}

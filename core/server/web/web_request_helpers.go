package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/metrics"
	"sumeru/core/orm"
)

// WebLogInput holds structured web log event fields.
type WebLogInput struct {
	Route         string
	Message       string
	Code          string
	Operation     string
	Status        string
	Err           error
	ContextFields map[string]interface{}
}

// WebLogEvent logs a structured web event using the applog contract.
func WebLogEvent(ctx context.Context, in WebLogInput) {
	contextFields := in.ContextFields
	if contextFields == nil {
		contextFields = map[string]interface{}{}
	}
	contextFields["route"] = in.Route

	event := applog.Event{
		Message:   in.Message,
		Code:      in.Code,
		Component: webLogComponent,
		Operation: in.Operation,
		Status:    in.Status,
		Context:   contextFields,
		Err:       in.Err,
	}
	emitWebLogEvent(ctx, event)
}

func emitWebLogEvent(ctx context.Context, event applog.Event) {
	switch {
	case event.Err != nil || event.Status == logStatusFailure:
		if event.Status == "" {
			event.Status = logStatusFailure
		}
		if event.Code != "" {
			applog.ErrorCode(ctx, event.Code, event.Message, event)
			return
		}
		applog.Error(ctx, event)
	case event.Status == logStatusPartial:
		if event.Code != "" {
			applog.WarnCode(ctx, event.Code, event.Message, event)
			return
		}
		applog.Warn(ctx, event)
	default:
		applog.Info(ctx, event)
	}
}

func WebLogf(ctx context.Context, route, format string, args ...interface{}) {
	if route = strings.TrimSpace(route); route == "" {
		route = webLogUnknownRoute
	}
	WebLogEvent(ctx, WebLogInput{
		Route: route, Message: fmt.Sprintf(format, args...),
		Operation: logOperationRequest, Status: logStatusSuccess,
	})
}

// WebLogNavigation emits an INFO-level audit event for successful navigation (menu, view, module, company).
func WebLogNavigation(ctx context.Context, route, operation, message string, fields map[string]interface{}) {
	WebLogEvent(ctx, WebLogInput{
		Route: route, Message: message, Operation: operation, Status: logStatusSuccess, ContextFields: fields,
	})
}

// SafePathNext returns a same-origin relative path, rejecting open-redirect targets.
func SafePathNext(rawURL, fallback string) string {
	return safeRedirectPath(rawURL, "/", fallback)
}

// SafeWebNext returns a path under /web, rejecting open-redirect targets.
func SafeWebNext(rawURL, fallback string) string {
	return safeRedirectPath(rawURL, "/web", fallback)
}

func safeRedirectPath(rawURL, requiredPrefix, fallback string) string {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" || !strings.HasPrefix(trimmed, requiredPrefix) || strings.HasPrefix(trimmed, "//") {
		return fallback
	}
	return trimmed
}

// redirectWithWebMessage redirects to a safe /web path with a flash message query parameter.
func redirectWithWebMessage(w http.ResponseWriter, r *http.Request, rawNext, message string) {
	nextPath := SafeWebNext(rawNext, homeRoute)
	redirectURL, err := urlWithQueryParam(nextPath, flashMessageParam, message)
	if err != nil {
		redirectURL = homeRoute + "?" + flashMessageParam + "=" + url.QueryEscape(message)
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func urlWithQueryParam(path, param, value string) (string, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set(param, value)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func RequirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	http.Error(w, methodNotAllowedMessage, http.StatusMethodNotAllowed)
	return false
}

func ParsePostForm(w http.ResponseWriter, r *http.Request) bool {
	const maxFormBytes = 1 << 20 // 1 MiB
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		http.Error(w, invalidFormMessage, http.StatusBadRequest)
		return false
	}
	return true
}

// requireLoginAndPOST checks login, POST method, form parsing, and CSRF (form body must be parsed first).
func requireLoginAndPOST(w http.ResponseWriter, r *http.Request) bool {
	return requireAuthenticatedPOST(w, r) &&
		ParsePostForm(w, r) &&
		validateSessionCSRF(w, r)
}

// requireLoginJSONPost checks login, POST method, and CSRF for JSON API handlers.
func requireLoginJSONPost(w http.ResponseWriter, r *http.Request) bool {
	return requireAuthenticatedPOST(w, r) && validateSessionCSRF(w, r)
}

// requireLoginMultipartPost checks login, POST method, bounded multipart parsing, and CSRF.
func requireLoginMultipartPost(w http.ResponseWriter, r *http.Request, maxBytes int64) bool {
	return requireAuthenticatedPOST(w, r) &&
		parseBoundedMultipartForm(w, r, maxBytes) &&
		validateSessionCSRF(w, r)
}

func requireAuthenticatedPOST(w http.ResponseWriter, r *http.Request) bool {
	return requireLogin(w, r) && RequirePOST(w, r)
}

func parseBoundedMultipartForm(w http.ResponseWriter, r *http.Request, maxBytes int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, invalidFormMessage, http.StatusBadRequest)
		return false
	}
	return true
}

func validateSessionCSRF(w http.ResponseWriter, r *http.Request) bool {
	if ValidateCSRF(r) {
		return true
	}
	metrics.Inc("sumeru_csrf_rejected_total")
	http.Error(w, invalidCSRFMessage, http.StatusForbidden)
	return false
}

func requireRegisteredModel(w http.ResponseWriter, modelName string) (orm.Model, bool) {
	model, registered := orm.Registry[modelName]
	if !registered || model == nil {
		http.Error(w, unknownModelMessage, http.StatusBadRequest)
		return nil, false
	}
	return model, true
}

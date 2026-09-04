package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/metrics"
	"sumeru/core/orm"
	"sumeru/core/server/api"
)

// APIHealthHandler is liveness: process is up (no dependency checks).
func APIHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSONOK(w)
}

// APIReadyHandler is readiness: PostgreSQL is reachable.
func APIReadyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if orm.DB == nil {
		api.WriteResponse(w, http.StatusServiceUnavailable, api.Fail(api.CodeInternalError, "database not initialized", nil))
		return
	}
	if err := orm.DB.Ping(); err != nil {
		api.WriteResponse(w, http.StatusServiceUnavailable, api.Fail(api.CodeInternalError, "database unavailable", nil))
		return
	}
	writeJSONOK(w)
}

// RPCJSONHandler is model RPC: POST JSON {"model","method","args","kwargs"} with session or API key auth.
func RPCJSONHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	metrics.Inc(rpcMetricRequests)
	defer func() { metrics.ObserveDuration(rpcMetricDuration, time.Since(start)) }()

	if r.Method != http.MethodPost {
		api.WriteResponse(w, http.StatusMethodNotAllowed, api.Fail(api.CodeMethodNotAllowed, "Method not allowed", nil))
		return
	}

	userID := AuthenticatedUserID(r)
	if userID <= 0 {
		api.WriteResponse(w, http.StatusUnauthorized, api.Fail(api.CodeUnauthorized, "Unauthorized", nil))
		return
	}
	// Cookie sessions require CSRF; API keys are not browser cookie auth.
	if SessionUserID(r) > 0 && !ValidateCSRF(r) {
		metrics.Inc("sumeru_csrf_rejected_total")
		api.WriteResponse(w, http.StatusForbidden, api.Fail(api.CodeAccessDenied, "Invalid CSRF token", nil))
		return
	}

	if !acceptsJSONContentType(r.Header.Get("Content-Type")) {
		api.WriteResponse(w, http.StatusUnsupportedMediaType, api.Fail(api.CodeUnsupportedMediaType, "Content-Type must be application/json", nil))
		return
	}

	requestBody, readOK := readBoundedRequestBody(r, maxRPCBodyBytes)
	if !readOK {
		api.WriteResponse(w, http.StatusBadRequest, api.Fail(api.CodeInvalidBody, "Could not read request body", nil))
		return
	}
	if int64(len(requestBody)) > maxRPCBodyBytes {
		api.WriteResponse(w, http.StatusRequestEntityTooLarge, api.Fail(api.CodePayloadTooLarge, "Request body too large", nil))
		return
	}

	ctx := rpcSecurityContext(r, userID)
	response, statusCode := api.Dispatch(ctx, requestBody)
	logRPCDispatch(ctx, requestBody, response, statusCode, start)
	api.WriteResponse(w, statusCode, response)
}

func acceptsJSONContentType(contentType string) bool {
	normalized := strings.TrimSpace(strings.ToLower(contentType))
	return normalized == "" || strings.HasPrefix(normalized, jsonContentTypePrefix)
}

func readBoundedRequestBody(r *http.Request, maxBytes int64) ([]byte, bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
	if err != nil {
		return nil, false
	}
	return body, true
}

func rpcSecurityContext(r *http.Request, userID int) context.Context {
	ctx := orm.ContextWithUID(r.Context(), userID)
	if companyID := orm.CompanyIDFromContext(r.Context()); companyID > 0 {
		return orm.ContextWithCompanyID(ctx, companyID)
	}
	return orm.ContextWithCompanyID(ctx, orm.ActiveCompanyIDForUser(ctx, userID))
}

func logRPCDispatch(ctx context.Context, requestBody []byte, response api.RPCResponse, statusCode int, start time.Time) {
	modelName, methodName := parseRPCCallMeta(requestBody)
	event := applog.Event{
		Component: "rpc",
		Operation: methodName,
		Duration:  time.Since(start),
		Context: map[string]interface{}{
			"resource":    modelName,
			"method":      methodName,
			"status_code": statusCode,
		},
	}
	if !response.OK && response.Error != nil {
		event.Message = "RPC call failed"
		event.Status = "failure"
		event.Context["error_code"] = response.Error.Code
		event.Context["error"] = response.Error.Message
		applog.Error(ctx, event)
		return
	}
	event.Message = "RPC call completed"
	event.Status = "success"
	applog.Info(ctx, event)
}

func parseRPCCallMeta(requestBody []byte) (modelName, methodName string) {
	var meta struct {
		Model  string `json:"model"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(requestBody, &meta); err != nil {
		return "", ""
	}
	return meta.Model, meta.Method
}

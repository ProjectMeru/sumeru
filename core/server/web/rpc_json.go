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

// RPCPrefixHandler routes POST /api/rpc/{model}/{method}, GET .../health, and POST auth.login.
func RPCPrefixHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	metrics.Inc(rpcMetricRequests)
	defer func() { metrics.ObserveDuration(rpcMetricDuration, time.Since(start)) }()

	if !strings.HasPrefix(r.URL.Path, apiRPCRoutePrefix) {
		api.WriteResponse(w, http.StatusNotFound, api.Fail(api.CodeNotFound, "Not found", nil))
		return
	}
	remainder := strings.TrimPrefix(r.URL.Path, apiRPCRoutePrefix)
	remainder = strings.Trim(remainder, "/")
	if remainder == "" || strings.Contains(remainder, "..") {
		api.WriteResponse(w, http.StatusNotFound, api.Fail(api.CodeNotFound, "Not found", nil))
		return
	}

	if ok, resp, status := validateRPCDatabaseHeader(r); !ok {
		api.WriteResponse(w, status, resp)
		return
	}
	companyID, ok, resp, status := resolveRPCTenantCompanyID(r)
	if !ok {
		api.WriteResponse(w, status, resp)
		return
	}

	if remainder == rpcAuthLoginSegment {
		if r.Method != http.MethodPost {
			api.WriteResponse(w, http.StatusMethodNotAllowed, api.Fail(api.CodeMethodNotAllowed, "Method not allowed", nil))
			return
		}
		if apiKeyFromRequest(r) != "" || SessionUserID(r) > 0 {
			api.WriteResponse(w, http.StatusBadRequest, api.Fail(api.CodeValidationError, "Authorization not allowed on auth.login", nil))
			return
		}
		body, readOK := readRPCBody(w, r)
		if !readOK {
			return
		}
		handleRPCAuthLogin(w, r, body, companyID)
		return
	}

	if strings.HasSuffix(remainder, "/health") {
		parts := strings.Split(remainder, "/")
		if len(parts) != 2 || parts[1] != "health" {
			api.WriteResponse(w, http.StatusNotFound, api.Fail(api.CodeNotFound, "Not found", nil))
			return
		}
		if r.Method != http.MethodGet {
			api.WriteResponse(w, http.StatusMethodNotAllowed, api.Fail(api.CodeMethodNotAllowed, "Method not allowed", nil))
			return
		}
		handleRPCModelHealth(w, r, parts[0], companyID)
		return
	}

	if r.Method != http.MethodPost {
		api.WriteResponse(w, http.StatusMethodNotAllowed, api.Fail(api.CodeMethodNotAllowed, "Method not allowed", nil))
		return
	}

	slash := strings.LastIndex(remainder, "/")
	if slash <= 0 || slash >= len(remainder)-1 {
		api.WriteResponse(w, http.StatusNotFound, api.Fail(api.CodeNotFound, "Not found", nil))
		return
	}
	modelName := remainder[:slash]
	methodName := remainder[slash+1:]
	if !api.PublicMethods[methodName] {
		api.WriteResponse(w, http.StatusNotFound, api.Fail(api.CodeMethodNotAllowed, "Method not allowed", map[string]interface{}{"method": methodName}))
		return
	}

	userID := AuthenticatedUserID(r)
	if userID <= 0 {
		msg, invalidKey := rpcUnauthorizedMessage(r)
		if invalidKey {
			api.WriteResponse(w, http.StatusUnauthorized, api.Fail(api.CodeUnauthorized, msg, nil))
		} else {
			api.WriteResponse(w, http.StatusUnauthorized, api.Fail(api.CodeUnauthorized, msg, nil))
		}
		return
	}
	if ok, resp, status := ensureUserAllowedTenant(r, userID, companyID); !ok {
		api.WriteResponse(w, status, resp)
		return
	}
	if AuthViaSession(r) && !ValidateCSRF(r) {
		metrics.Inc("sumeru_csrf_rejected_total")
		api.WriteResponse(w, http.StatusForbidden, api.Fail(api.CodeAccessDenied, "Invalid CSRF token", nil))
		return
	}

	body, readOK := readRPCBody(w, r)
	if !readOK {
		return
	}

	ctx := rpcSecurityContext(r, userID, companyID)
	response, statusCode := api.DispatchPath(ctx, modelName, methodName, body)
	logRPCDispatch(ctx, modelName, methodName, response, statusCode, start)
	api.WriteResponse(w, statusCode, response)
}

// RPCJSONHandler is kept as an alias for tests that still call the prefix handler.
func RPCJSONHandler(w http.ResponseWriter, r *http.Request) {
	RPCPrefixHandler(w, r)
}

func readRPCBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	if !acceptsJSONContentType(r.Header.Get("Content-Type")) {
		api.WriteResponse(w, http.StatusUnsupportedMediaType, api.Fail(api.CodeUnsupportedMediaType, "Content-Type must be application/json", nil))
		return nil, false
	}
	requestBody, readOK := readBoundedRequestBody(r, maxRPCBodyBytes)
	if !readOK {
		api.WriteResponse(w, http.StatusBadRequest, api.Fail(api.CodeInvalidBody, "Could not read request body", nil))
		return nil, false
	}
	if int64(len(requestBody)) > maxRPCBodyBytes {
		api.WriteResponse(w, http.StatusRequestEntityTooLarge, api.Fail(api.CodePayloadTooLarge, "Request body too large", nil))
		return nil, false
	}
	return requestBody, true
}

func handleRPCModelHealth(w http.ResponseWriter, r *http.Request, modelName string, companyID int64) {
	userID := AuthenticatedUserID(r)
	if userID <= 0 {
		msg, _ := rpcUnauthorizedMessage(r)
		api.WriteResponse(w, http.StatusUnauthorized, api.Fail(api.CodeUnauthorized, msg, nil))
		return
	}
	if ok, resp, status := ensureUserAllowedTenant(r, userID, companyID); !ok {
		api.WriteResponse(w, status, resp)
		return
	}
	_, registered := orm.Registry[modelName]
	methods := make([]string, 0, len(api.PublicMethods))
	for name, allowed := range api.PublicMethods {
		if allowed {
			methods = append(methods, name)
		}
	}
	if !registered {
		api.WriteResponse(w, http.StatusNotFound, api.Fail(api.CodeModelNotFound, "unknown model", map[string]interface{}{"model": modelName}))
		return
	}
	data := map[string]interface{}{
		"model":       modelName,
		"registered":  true,
		"rpc_methods": methods,
	}
	resp := api.SuccessWithMessage(data, "Model health OK")
	api.WriteResponse(w, http.StatusOK, resp)
}

func acceptsJSONContentType(contentType string) bool {
	normalized := strings.TrimSpace(strings.ToLower(contentType))
	return strings.HasPrefix(normalized, jsonContentTypePrefix)
}

func readBoundedRequestBody(r *http.Request, maxBytes int64) ([]byte, bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
	if err != nil {
		return nil, false
	}
	return body, true
}

func rpcSecurityContext(r *http.Request, userID int, companyID int64) context.Context {
	ctx := orm.ContextWithUID(r.Context(), userID)
	if companyID > 0 {
		return orm.ContextWithCompanyID(ctx, companyID)
	}
	if cid := orm.CompanyIDFromContext(r.Context()); cid > 0 {
		return orm.ContextWithCompanyID(ctx, cid)
	}
	return orm.ContextWithCompanyID(ctx, orm.ActiveCompanyIDForUser(ctx, userID))
}

func logRPCDispatch(ctx context.Context, modelName, methodName string, response api.APIResponse, statusCode int, start time.Time) {
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
	if !response.Success && response.Error != nil {
		event.Message = response.Error.Message
		if event.Message == "" {
			event.Message = response.Message
		}
		if event.Message == "" {
			event.Message = "RPC call failed"
		}
		event.Status = "failure"
		if response.Error.Context != nil {
			if code, ok := response.Error.Context["code"].(string); ok {
				event.Code = code
			}
		}
		if event.Code == "" {
			event.Code = response.Error.Code
		}
		event.Context["error"] = event.Message
		applog.ErrorCode(ctx, event.Code, event.Message, event)
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


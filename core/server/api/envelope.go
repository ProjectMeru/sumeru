package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIError is the structured error object in /api/rpc responses.
type APIError struct {
	Code      string                 `json:"code,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Arguments []any                  `json:"arguments,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Debug     any                    `json:"debug,omitempty"`
}

// APIResponse is the standard envelope for /api/rpc.
type APIResponse struct {
	Success    bool        `json:"success"`
	StatusCode int         `json:"status_code"`
	Timestamp  time.Time   `json:"timestamp"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Error      *APIError   `json:"error"`
}

// RPCResponse is an alias kept for transitional references in tests.
type RPCResponse = APIResponse

// RPCError is an alias kept for transitional references in tests.
type RPCError = APIError

func apiNow() time.Time {
	return time.Now()
}

// Success builds a successful APIResponse.
func Success(result interface{}) APIResponse {
	return SuccessWithMessage(result, "")
}

// SuccessWithMessage builds a successful APIResponse with a human message.
func SuccessWithMessage(result interface{}, message string) APIResponse {
	if message == "" {
		message = "OK"
	}
	return APIResponse{
		Success:    true,
		StatusCode: http.StatusOK,
		Timestamp:  apiNow(),
		Message:    message,
		Data:       result,
		Error:      nil,
	}
}

// Fail builds a failed APIResponse for model RPC (machine code in context).
func Fail(code, message string, details map[string]interface{}) APIResponse {
	status := statusForCode(code)
	ctx := details
	if ctx == nil {
		ctx = map[string]interface{}{}
	}
	if _, ok := ctx["code"]; !ok && code != "" {
		ctx["code"] = code
	}
	return APIResponse{
		Success:    false,
		StatusCode: status,
		Timestamp:  apiNow(),
		Message:    message,
		Data:       nil,
		Error: &APIError{
			Code:    code,
			Name:    rpcErrorName(code),
			Message: message,
			Context: ctx,
		},
	}
}

// FailWithStatus builds Fail with an explicit HTTP status on the envelope.
func FailWithStatus(status int, code, message string, details map[string]interface{}) APIResponse {
	resp := Fail(code, message, details)
	resp.StatusCode = status
	return resp
}

// AuthFail builds an authentication-style error envelope.
func AuthFail(status int, code, typ, message, debug string) APIResponse {
	return APIResponse{
		Success:    false,
		StatusCode: status,
		Timestamp:  apiNow(),
		Message:    message,
		Data:       nil,
		Error: &APIError{
			Code:    code,
			Type:    typ,
			Context: map[string]interface{}{},
			Debug:   debug,
		},
	}
}

func rpcErrorName(code string) string {
	if code == "" {
		return "sumeru.api.Error"
	}
	return fmt.Sprintf("sumeru.api.%s", code)
}

// WriteResponse writes an APIResponse with the given HTTP status.
func WriteResponse(w http.ResponseWriter, status int, resp APIResponse) {
	if resp.StatusCode == 0 {
		resp.StatusCode = status
	}
	if resp.Timestamp.IsZero() {
		resp.Timestamp = apiNow()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ResponseFromError maps an error to an APIResponse and HTTP status.
func ResponseFromError(err error) (APIResponse, int) {
	if err == nil {
		return Success(nil), http.StatusOK
	}
	if ce, ok := err.(*codedError); ok {
		return Fail(ce.code, ce.msg, ce.details), statusForCode(ce.code)
	}
	code, details := Classify(err)
	return Fail(code, err.Error(), details), statusForCode(code)
}

// Dispatch executes a legacy flat-body RPC call.
func Dispatch(ctx context.Context, body []byte) (APIResponse, int) {
	result, err := dispatchRPC(ctx, body)
	if err != nil {
		return ResponseFromError(err)
	}
	return Success(result), http.StatusOK
}

// DispatchPath executes POST /api/rpc/{model}/{method} with a method-shaped body.
func DispatchPath(ctx context.Context, model, method string, body []byte) (APIResponse, int) {
	built, err := BuildRPCRequestBody(method, body)
	if err != nil {
		if ce, ok := err.(*codedError); ok {
			return Fail(ce.code, ce.msg, ce.details), statusForCode(ce.code)
		}
		return Fail(CodeInvalidJSON, err.Error(), nil), http.StatusBadRequest
	}
	in := rpcRequest{Model: model, Method: method, Args: built.args, Kwargs: built.kwargs}
	payload, err := json.Marshal(in)
	if err != nil {
		return Fail(CodeInternalError, "could not build rpc request", nil), http.StatusInternalServerError
	}
	result, err := dispatchRPC(ctx, payload)
	if err != nil {
		resp, status := ResponseFromError(err)
		resp.Message = rpcFailureMessage(model, method, err)
		return resp, status
	}
	msg := rpcSuccessMessage(model, method)
	return SuccessWithMessage(result, msg), http.StatusOK
}

func rpcSuccessMessage(model, method string) string {
	switch method {
	case "search_read", "search":
		return "Records retrieved successfully"
	case "read":
		return "Records retrieved successfully"
	case "create", "create_many":
		return "Record created successfully"
	case "write", "write_many":
		return "Record updated successfully"
	case "unlink", "unlink_many":
		return "Record deleted successfully"
	default:
		return fmt.Sprintf("%s.%s completed successfully", model, method)
	}
}

func rpcFailureMessage(model, method string, err error) string {
	if err != nil && err.Error() != "" {
		return err.Error()
	}
	return fmt.Sprintf("%s.%s failed", model, method)
}

func statusForCode(code string) int {
	switch code {
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeValidationError, CodeInvalidJSON, CodeInvalidArgs, CodeInvalidBody:
		return http.StatusBadRequest
	case CodeUnsupportedMediaType:
		return http.StatusUnsupportedMediaType
	case CodePayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeMethodNotAllowed, CodeAccessDenied:
		return http.StatusForbidden
	case CodeModelNotFound, CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

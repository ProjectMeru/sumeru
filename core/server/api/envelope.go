package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// RPCError is the structured error object in POST /api/rpc responses.
type RPCError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// RPCResponse is the standard envelope for POST /api/rpc.
type RPCResponse struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result"`
	Error  *RPCError   `json:"error"`
}

// Success builds a successful RPCResponse.
func Success(result interface{}) RPCResponse {
	return RPCResponse{OK: true, Result: result, Error: nil}
}

// Fail builds a failed RPCResponse.
func Fail(code, message string, details map[string]interface{}) RPCResponse {
	return RPCResponse{
		OK:     false,
		Result: nil,
		Error:  &RPCError{Code: code, Message: message, Details: details},
	}
}

// WriteResponse writes an RPCResponse with the given HTTP status.
func WriteResponse(w http.ResponseWriter, status int, resp RPCResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ResponseFromError maps an error to an RPCResponse and HTTP status.
func ResponseFromError(err error) (RPCResponse, int) {
	if err == nil {
		return Success(nil), http.StatusOK
	}
	if ce, ok := err.(*codedError); ok {
		return Fail(ce.code, ce.msg, ce.details), statusForCode(ce.code)
	}
	code, details := Classify(err)
	return Fail(code, err.Error(), details), statusForCode(code)
}

// Dispatch executes an RPC call and returns the standard envelope and HTTP status.
func Dispatch(ctx context.Context, body []byte) (RPCResponse, int) {
	result, err := dispatchRPC(ctx, body)
	if err != nil {
		return ResponseFromError(err)
	}
	return Success(result), http.StatusOK
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

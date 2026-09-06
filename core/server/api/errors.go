package api

import (
	"strings"

	"sumeru/core/orm"
)

const (
	CodeInvalidJSON          = "INVALID_JSON"
	CodeInvalidArgs          = "INVALID_ARGS"
	CodeInvalidBody          = "INVALID_BODY"
	CodeValidationError      = "VALIDATION_ERROR" // errcode.ValidationError
	CodeModelNotFound        = "MODEL_NOT_FOUND"
	CodeNotFound             = "NOT_FOUND" // errcode.NotFound
	CodeMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	CodeUnauthorized         = "UNAUTHORIZED" // errcode.Unauthorized
	CodeAccessDenied         = "ACCESS_DENIED" // errcode.AccessDenied
	CodePayloadTooLarge      = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	CodeInternalError        = "INTERNAL_ERROR" // errcode.InternalError
)

type codedError struct {
	code    string
	msg     string
	details map[string]interface{}
}

func (e *codedError) Error() string {
	return e.msg
}

func newRPCError(code, msg string, details map[string]interface{}) error {
	if details == nil {
		details = map[string]interface{}{}
	}
	return &codedError{code: code, msg: msg, details: details}
}

// Classify maps a generic error message to a standard RPC error code and optional details.
func Classify(err error) (code string, details map[string]interface{}) {
	if err == nil {
		return CodeInternalError, nil
	}
	if orm.IsAccessDenied(err) {
		return CodeAccessDenied, map[string]interface{}{}
	}
	msg := strings.ToLower(err.Error())
	details = map[string]interface{}{}

	switch {
	case strings.Contains(msg, "invalid json"), strings.Contains(msg, "invalid params"):
		return CodeInvalidJSON, details
	case strings.Contains(msg, "empty body"):
		return CodeInvalidJSON, details
	case strings.Contains(msg, "args:"), strings.Contains(msg, "requires args"), strings.Contains(msg, "kwargs must"):
		return CodeInvalidArgs, details
	case strings.Contains(msg, "model is required"):
		details["field"] = "model"
		return CodeValidationError, details
	case strings.Contains(msg, "method is required"):
		details["field"] = "method"
		return CodeValidationError, details
	case strings.Contains(msg, "unknown model"), strings.Contains(msg, "not registered"):
		return CodeModelNotFound, details
	case strings.Contains(msg, "not a public rpc method"), strings.Contains(msg, "unsupported method"):
		return CodeMethodNotAllowed, details
	case strings.Contains(msg, "authentication required"):
		return CodeUnauthorized, details
	case strings.Contains(msg, "access denied"):
		return CodeAccessDenied, details
	case strings.Contains(msg, "record(s) not found"), strings.Contains(msg, "not found"):
		return CodeNotFound, details
	default:
		return CodeInternalError, details
	}
}

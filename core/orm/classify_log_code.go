package orm

import (
	"errors"
	"strings"

	"sumeru/core/errcode"
)

// ClassifyLogCode maps an error to a stable applog/errcode machine id.
func ClassifyLogCode(err error) string {
	if err == nil {
		return errcode.InternalError
	}
	if IsAccessDenied(err) || IsRecordRuleFailed(err) {
		return errcode.AccessDenied
	}
	var fieldErr *FieldValidationError
	if errors.As(err, &fieldErr) {
		return errcode.ValidationError
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "record(s) not found"), strings.Contains(msg, "not found"):
		return errcode.RecordNotFound
	case strings.Contains(msg, "validation"), strings.Contains(msg, "invalid"), strings.Contains(msg, "required"):
		return errcode.ValidationError
	default:
		return errcode.InternalError
	}
}

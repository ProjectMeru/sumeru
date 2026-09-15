package orm

import (
	"errors"
	"fmt"
)

// AccessDeniedError indicates the user lacks model-level ACL permission.
type AccessDeniedError struct {
	Model     string
	Operation string
}

func (e *AccessDeniedError) Error() string {
	if e == nil {
		return "access denied"
	}
	if e.Model != "" && e.Operation != "" {
		return "access denied on " + e.Model + " for operation " + e.Operation
	}
	if e.Model != "" {
		return "access denied on " + e.Model
	}
	return "access denied"
}

// RecordRuleError indicates a record rule blocked the operation.
type RecordRuleError struct {
	Model string
}

func (e *RecordRuleError) Error() string {
	if e == nil || e.Model == "" {
		return "record rule failed"
	}
	return "record rule failed for model " + e.Model
}

func IsAccessDenied(err error) bool {
	var target *AccessDeniedError
	return errors.As(err, &target)
}

func IsRecordRuleFailed(err error) bool {
	var target *RecordRuleError
	return errors.As(err, &target)
}

// SmuggledBypassError indicates bypass was set on an authenticated user context outside WithElevated.
type SmuggledBypassError struct {
	UID int
}

func (e *SmuggledBypassError) Error() string {
	if e == nil || e.UID <= 0 {
		return "security bypass not allowed for authenticated user"
	}
	return fmt.Sprintf("security bypass not allowed for user %d", e.UID)
}

func IsSmuggledBypass(err error) bool {
	var target *SmuggledBypassError
	return errors.As(err, &target)
}

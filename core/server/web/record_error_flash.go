package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

func userFacingRecordError(operation, model string, err error) (title, body, details string, fieldErrors []string) {
	if err == nil {
		return "Error", "An unknown error occurred.", "", nil
	}
	details = err.Error()

	var fve *orm.FieldValidationError
	if errors.As(err, &fve) && fve != nil {
		switch operation {
		case "record_save":
			title = "Save failed"
		default:
			title = "Validation failed"
		}
		body = fve.Error()
		if fve.Field != "" {
			fieldErrors = []string{fve.Field}
		}
		return title, body, details, fieldErrors
	}

	switch operation {
	case "record_save":
		title = "Save failed"
	case "record_delete":
		title = "Delete failed"
	case "record_chatter":
		title = "Message not posted"
	case "object_action":
		title = "Action failed"
	default:
		title = "Request failed"
	}

	if orm.IsRecordRuleFailed(err) {
		if operation == "record_delete" {
			body = "You don't have permission to delete this record."
		} else {
			body = "This record could not be saved with your access rights."
		}
		return title, body, details, fieldErrors
	}
	if orm.IsAccessDenied(err) {
		body = "You don't have permission to perform this action."
		return title, body, details, fieldErrors
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "required") || strings.HasPrefix(msg, "field ") {
		body = friendlyRequiredMessage(model, err)
		fieldErrors = inferFieldErrorsFromMessage(err.Error())
		return title, body, details, fieldErrors
	}
	if strings.Contains(msg, "unknown field") {
		body = "One or more submitted fields are not valid for this form."
		return title, body, details, fieldErrors
	}
	if strings.Contains(msg, "no valid fields") {
		body = err.Error()
		return title, body, details, fieldErrors
	}

	if model != "" {
		body = fmt.Sprintf("%s for %s. Review the details below or contact your administrator.", title, model)
	} else {
		body = fmt.Sprintf("%s. Review the details below or contact your administrator.", title)
	}
	return title, body, details, fieldErrors
}

func friendlyRequiredMessage(model string, err error) string {
	msg := err.Error()
	if strings.Contains(msg, `required field "`) {
		start := strings.Index(msg, `"`) + 1
		end := strings.Index(msg[start:], `"`)
		if end > 0 {
			fieldName := msg[start : start+end]
			if label := fieldLabel(model, fieldName); label != "" {
				return fmt.Sprintf("%s is required.", label)
			}
		}
	}
	return msg
}

func inferFieldErrorsFromMessage(msg string) []string {
	if strings.Contains(msg, `required field "`) {
		start := strings.Index(msg, `"`) + 1
		end := strings.Index(msg[start:], `"`)
		if end > 0 {
			return []string{msg[start : start+end]}
		}
	}
	if strings.Contains(strings.ToLower(msg), " is required.") {
		return nil
	}
	return nil
}

func fieldLabel(modelName, fieldName string) string {
	inst, ok := orm.Registry[modelName]
	if !ok || inst == nil {
		return fieldName
	}
	for _, f := range inst.Fields() {
		if f.Name == fieldName {
			if s := strings.TrimSpace(f.String); s != "" {
				return s
			}
			return fieldName
		}
	}
	return fieldName
}

func redirectRecordError(w http.ResponseWriter, r *http.Request, nextURL, operation, model string, err error) {
	ctx := r.Context()
	title, body, details, fieldErrors := userFacingRecordError(operation, model, err)
	WebLogEvent(ctx, WebLogInput{
		Route: operationRoute(operation), Message: body,
		Operation: operation, Status: logStatusFailure, Err: err,
		ContextFields: map[string]interface{}{"model": model},
	})
	applog.DebugMsg(ctx, webLogComponent, operation, "record POST failed", map[string]interface{}{
		"model": model,
		"error": err.Error(),
	})
	SetRecordErrorFlash(w, PageFlash{
		Kind:        "error",
		Title:       title,
		Body:        body,
		Details:     details,
		FieldErrors: fieldErrors,
	})
	redirectURL := appendFieldErrorsToURL(SafeWebNext(nextURL, homeRoute), fieldErrors)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func appendFieldErrorsToURL(rawURL string, fieldErrors []string) string {
	if len(fieldErrors) == 0 {
		return rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	query.Set(fieldErrorsParam, strings.Join(fieldErrors, ","))
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func operationRoute(operation string) string {
	switch operation {
	case "record_save", "record_delete":
		return workspaceRoute
	case "record_chatter":
		return chatterPostRoute
	case "object_action":
		return apiRPCRoute
	default:
		return webLogUnknownRoute
	}
}

func ensureFormEditRedirectURL(rawNext string, clearRecordID bool) string {
	nextPath := SafeWebNext(rawNext, homeRoute)
	parsed, err := url.Parse(nextPath)
	if err != nil {
		return nextPath
	}
	query := parsed.Query()
	query.Set(workspaceEditParam, workspaceEditEnabledValue)
	if clearRecordID {
		query.Del(workspaceRecordIDParam)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

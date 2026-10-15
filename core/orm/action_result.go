package orm

import (
	"net/url"
	"strconv"
	"strings"
)

const actionResultClose = "close"
const actionTargetDialog = "dialog"
const actionTargetCurrent = "current"

// ActionOpen describes opening a form (typically a wizard) after an object action.
type ActionOpen struct {
	Model    string `json:"model,omitempty"`
	ActionID int    `json:"actionId,omitempty"`
	ViewType string `json:"viewType,omitempty"`
	RecordID int    `json:"recordId,omitempty"`
	Target   string `json:"target,omitempty"` // dialog | current
}

// ActionResult is the JSON contract returned by RPC call (object buttons).
type ActionResult struct {
	Redirect string      `json:"redirect,omitempty"`
	Open     *ActionOpen `json:"open,omitempty"`
	Close    bool        `json:"close,omitempty"`
}

// ActionOpenURL builds a wizard workspace URL (model + id, dialog target).
func ActionOpenURL(model string, id int) string {
	model = strings.TrimSpace(model)
	if model == "" || 0 >= id {
		return ""
	}
	q := url.Values{}
	q.Set("model", model)
	q.Set("id", strconv.Itoa(id))
	q.Set("view_type", "form")
	q.Set("target", actionTargetDialog)
	return "/web?" + q.Encode()
}

// EncodeActionResult maps a legacy object-action redirect string into the RPC result.
// Empty → true (stay). "close" → {close:true}. /web?model=&id= → {open:...}. Else {redirect}.
func EncodeActionResult(redirect string) interface{} {
	redirect = strings.TrimSpace(redirect)
	if redirect == "" {
		return true
	}
	if strings.EqualFold(redirect, actionResultClose) {
		return map[string]interface{}{"close": true}
	}
	if open := parseWorkspaceOpen(redirect); open != nil {
		return map[string]interface{}{"open": open}
	}
	return map[string]interface{}{"redirect": redirect}
}

func parseWorkspaceOpen(redirect string) map[string]interface{} {
	raw := redirect
	if strings.HasPrefix(raw, "/web?") {
		raw = strings.TrimPrefix(raw, "/web")
	} else if !strings.HasPrefix(raw, "?") {
		return nil
	}
	q, err := url.ParseQuery(strings.TrimPrefix(raw, "?"))
	if err != nil {
		return nil
	}
	model := strings.TrimSpace(q.Get("model"))
	id, _ := strconv.Atoi(strings.TrimSpace(q.Get("id")))
	if model == "" || 0 >= id {
		return nil
	}
	viewType := strings.TrimSpace(q.Get("view_type"))
	if viewType == "" {
		viewType = "form"
	}
	target := strings.TrimSpace(q.Get("target"))
	if target == "" {
		target = actionTargetDialog
	}
	if target != actionTargetDialog && target != actionTargetCurrent {
		target = actionTargetDialog
	}
	open := map[string]interface{}{
		"model":    model,
		"viewType": viewType,
		"recordId": id,
		"target":   target,
	}
	if actionID, err := strconv.Atoi(strings.TrimSpace(q.Get("action"))); err == nil && actionID > 0 {
		open["actionId"] = actionID
	}
	return open
}

// ActionCloseToken is the redirect string handlers return to close a wizard dialog.
func ActionCloseToken() string {
	return actionResultClose
}

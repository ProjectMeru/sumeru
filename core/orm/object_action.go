package orm

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ObjectActionFunc handles a named object button (type="object") on a record.
// Return a non-empty Redirect to send the browser there after success.
type ObjectActionFunc func(ctx context.Context, model string, id int, vals map[string]string) (redirect string, err error)

var (
	objectActionMu sync.RWMutex
	objectActions  = map[string]map[string]ObjectActionFunc{} // model → method → fn
)

// RegisterObjectAction registers a handler for model + method (e.g. account.move / action_post).
// Later registration for the same pair replaces the earlier handler.
func RegisterObjectAction(model, method string, fn ObjectActionFunc) {
	model = strings.TrimSpace(model)
	method = strings.TrimSpace(method)
	if model == "" || method == "" || fn == nil {
		return
	}
	objectActionMu.Lock()
	defer objectActionMu.Unlock()
	if objectActions[model] == nil {
		objectActions[model] = map[string]ObjectActionFunc{}
	}
	objectActions[model][method] = fn
}

// RunObjectAction invokes a registered object action.
func RunObjectAction(ctx context.Context, model string, id int, method string, vals map[string]string) (string, error) {
	model = strings.TrimSpace(model)
	method = strings.TrimSpace(method)
	if model == "" || method == "" {
		return "", fmt.Errorf("model and method are required")
	}
	if id <= 0 {
		return "", fmt.Errorf("invalid record id")
	}
	if err := RejectSmuggledUserBypass(ctx); err != nil {
		return "", err
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, model, "write"); err != nil {
		return "", err
	}
	if _, err := SearchOne(ctx, model, map[string]interface{}{"id": id}); err != nil {
		return "", fmt.Errorf("record not found")
	}
	objectActionMu.RLock()
	var fn ObjectActionFunc
	if byMethod := objectActions[model]; byMethod != nil {
		fn = byMethod[method]
	}
	objectActionMu.RUnlock()
	if fn == nil {
		return "", fmt.Errorf("unknown object action %s.%s", model, method)
	}
	return fn(ctx, model, id, vals)
}

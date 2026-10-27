package orm

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ObjectActionFunc handles a named object button (type="object") on one record.
// Return a non-empty Redirect to send the browser there after success.
type ObjectActionFunc func(ctx context.Context, model string, id int, vals map[string]string) (redirect string, err error)

// ObjectActionIDsFunc handles the same button for every id in one call.
type ObjectActionIDsFunc func(ctx context.Context, model string, ids []int, vals map[string]string) (redirect string, err error)

var (
	objectActionMu sync.RWMutex
	objectActions  = map[string]map[string]ObjectActionIDsFunc{} // model → method → fn
)

// RegisterObjectAction registers a handler for model + method.
// Later registration for the same pair replaces the earlier handler.
// A multi-id call runs the handler once per id.
func RegisterObjectAction(model, method string, fn ObjectActionFunc) {
	if fn == nil {
		return
	}
	RegisterObjectActionIDs(model, method, func(ctx context.Context, model string, ids []int, vals map[string]string) (string, error) {
		var redirect string
		for _, id := range ids {
			next, err := fn(ctx, model, id, vals)
			if err != nil {
				return "", err
			}
			if next != "" {
				redirect = next
			}
		}
		return redirect, nil
	})
}

// RegisterObjectActionIDs registers a handler that receives the full id list.
// Later registration for the same pair replaces the earlier handler.
func RegisterObjectActionIDs(model, method string, fn ObjectActionIDsFunc) {
	model = strings.TrimSpace(model)
	method = strings.TrimSpace(method)
	if model == "" || method == "" || fn == nil {
		return
	}
	objectActionMu.Lock()
	defer objectActionMu.Unlock()
	if objectActions[model] == nil {
		objectActions[model] = map[string]ObjectActionIDsFunc{}
	}
	objectActions[model][method] = fn
}

// RunObjectAction invokes a registered object action for one id.
func RunObjectAction(ctx context.Context, model string, id int, method string, vals map[string]string) (string, error) {
	return RunObjectActionIDs(ctx, model, []int{id}, method, vals)
}

// RunObjectActionIDs invokes a registered object action for every id.
// Write access is checked once. Each id must exist.
func RunObjectActionIDs(ctx context.Context, model string, ids []int, method string, vals map[string]string) (string, error) {
	model = strings.TrimSpace(model)
	method = strings.TrimSpace(method)
	if model == "" || method == "" {
		return "", fmt.Errorf("model and method are required")
	}
	if 0 == len(ids) {
		return "", fmt.Errorf("invalid record id")
	}
	for _, id := range ids {
		if 0 >= id {
			return "", fmt.Errorf("invalid record id")
		}
	}
	if err := RejectSmuggledUserBypass(ctx); err != nil {
		return "", err
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, model, "write"); err != nil {
		return "", err
	}
	for _, id := range ids {
		if _, err := SearchOne(ctx, model, map[string]interface{}{"id": id}); err != nil {
			return "", fmt.Errorf("record not found")
		}
	}
	objectActionMu.RLock()
	var fn ObjectActionIDsFunc
	if byMethod := objectActions[model]; byMethod != nil {
		fn = byMethod[method]
	}
	objectActionMu.RUnlock()
	if fn == nil {
		return "", fmt.Errorf("unknown object action %s.%s", model, method)
	}
	return fn(ctx, model, ids, vals)
}

package web

import (
	"context"
	"encoding/json"
	"net/http"

	"sumeru/core/orm"
)

func writeJSON(w http.ResponseWriter, ctx context.Context, route string, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil && ctx != nil && route != "" {
		WebLogEvent(ctx, WebLogInput{
			Route: route, Message: "Failed to encode JSON response",
			Operation: "write", Status: "partial", Err: err,
		})
	}
}

func writeJSONOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func writeJSONResponse(w http.ResponseWriter, v interface{}) {
	writeJSON(w, nil, "", v)
}

// resolveRequestUID returns SecurityUID from context, falling back to the session cookie.
func resolveRequestUID(r *http.Request) int {
	uid := orm.SecurityUID(r.Context())
	if uid <= 0 {
		uid = SessionUserID(r)
	}
	return uid
}

func requireSystemAdmin(w http.ResponseWriter, r *http.Request, redirectOnDeny bool) bool {
	ctx := r.Context()
	uid := resolveRequestUID(r)
	if orm.UserHasGroupXML(ctx, uid, groupSystemXML) {
		return true
	}
	if redirectOnDeny {
		http.Redirect(w, r, homeRoute, http.StatusFound)
	} else {
		http.Error(w, forbiddenMessage, http.StatusForbidden)
	}
	return false
}

func requireModelAccess(w http.ResponseWriter, r *http.Request, model, perm string) bool {
	if err := orm.CheckModelAccess(r.Context(), orm.SecurityUID(r.Context()), model, perm); err != nil {
		WebLogEvent(r.Context(), WebLogInput{
			Route: r.URL.Path, Message: "Model access denied",
			Operation: "access", Status: "failure", Err: err,
			ContextFields: map[string]interface{}{"resource": model, "permission": perm},
		})
		http.Error(w, forbiddenMessage, http.StatusForbidden)
		return false
	}
	return true
}

func requireMenuAccess(w http.ResponseWriter, r *http.Request, menuXMLID string) bool {
	ctx := r.Context()
	uid := resolveRequestUID(r)
	menuID, _, err := orm.ResolveXmlId(ctx, menuXMLID)
	if err == nil && menuID > 0 {
		if rec, err := orm.SearchOne(ctx, sysMenuModel, map[string]interface{}{"id": menuID}); err == nil {
			if userMayAccessMenuRecord(ctx, uid, rec) {
				return true
			}
		}
	}
	if orm.UserHasGroupXML(ctx, uid, groupSystemXML) {
		return true
	}
	http.Error(w, forbiddenMessage, http.StatusForbidden)
	return false
}

func userMayAccessMenuRecord(ctx context.Context, uid int, menuRecord map[string]interface{}) bool {
	return orm.UserMayAccessMenu(ctx, uid, orm.AsString(menuRecord["access_groups"]))
}

func userMayAccessMenuByID(ctx context.Context, uid int, menuID int) bool {
	rec, err := orm.SearchOne(ctx, sysMenuModel, map[string]interface{}{"id": menuID})
	if err != nil {
		return false
	}
	return userMayAccessMenuRecord(ctx, uid, rec)
}

package web

import (
	"context"
	"net/http"

	"sumeru/core/orm"
)

func requirePortalUser(w http.ResponseWriter, r *http.Request) bool {
	if !requireLogin(w, r) {
		return false
	}
	uid := SessionUserID(r)
	if uid <= 0 {
		return false
	}
	if !orm.UserIsPortalOnly(r.Context(), uid) && !orm.UserHasGroupXML(r.Context(), uid, "base.group_user") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func redirectPortalUserFromWeb(w http.ResponseWriter, r *http.Request, uid int) bool {
	if uid <= 0 || !orm.UserIsPortalOnly(r.Context(), uid) {
		return false
	}
	http.Redirect(w, r, portalHomeRoute, http.StatusFound)
	return true
}

func postLoginDestination(ctx context.Context, userID int, requestedNext string) string {
	if orm.UserIsPortalOnly(ctx, userID) {
		if stringsHasPortalPrefix(requestedNext) {
			return SafePathNext(requestedNext, portalHomeRoute)
		}
		return portalHomeRoute
	}
	return SafePathNext(requestedNext, homeRoute)
}

func stringsHasPortalPrefix(path string) bool {
	return len(path) >= len(portalRoutePrefix) && path[:len(portalRoutePrefix)] == portalRoutePrefix
}

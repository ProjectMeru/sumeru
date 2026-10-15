package web

import (
	"net/http"
	"strconv"

	"sumeru/core/orm"
)

const userSecurityPostRoute = "/web/user/security-post"

func registerUserSecurityRoutes() {
	registerSession(http.MethodPost, userSecurityPostRoute, UserSecurityPostHandler)
}

// UserSecurityPostHandler applies password, groups, and companies after core.user save.
func UserSecurityPostHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	actor := orm.SecurityUID(ctx)
	userID, _ := strconv.Atoi(r.PostForm.Get("user_id"))
	if 0 >= userID {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}
	if !ValidateCSRF(r) {
		http.Error(w, "invalid csrf", http.StatusForbidden)
		return
	}
	if err := orm.ApplyUserSecurityPostErr(ctx, actor, userID, r.PostForm); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

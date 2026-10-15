package web

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"sumeru/addons/im"
	"sumeru/core/orm"
)

const (
	swcDirectUsersRoute         = "/web/swc/direct/users"
	swcDirectConversationsRoute = "/web/swc/direct/conversations"
	swcDirectThreadRoute        = "/web/swc/direct/thread"
	swcDirectPostRoute          = "/web/swc/direct/post"
	swcDirectUploadRoute        = "/web/swc/direct/upload"
	swcDirectStatusRoute        = "/web/swc/direct/status"
)

func registerSwcDirectRoutes() {
	registerSession(http.MethodGet, swcDirectStatusRoute, SwcDirectStatusHandler)
	registerSession(http.MethodGet, swcDirectUsersRoute, SwcDirectUsersHandler)
	registerSession(http.MethodGet, swcDirectConversationsRoute, SwcDirectConversationsHandler)
	registerSession(http.MethodGet, swcDirectThreadRoute, SwcDirectThreadHandler)
	registerSession(http.MethodPost, swcDirectPostRoute, SwcDirectPostHandler)
	registerSession(http.MethodPost, swcDirectUploadRoute, SwcDirectUploadHandler)
}

// SwcDirectStatusHandler GET /web/swc/direct/status — IM enabled flag, unread count, company settings link.
func SwcDirectStatusHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	uid := AuthenticatedUserID(r)
	companyID := int(orm.CompanyIDFromContext(ctx))
	if 0 >= companyID {
		companyID = int(orm.ActiveCompanyIDForUser(ctx, uid))
	}
	unread := 0
	if im.IMEnabled(ctx) {
		if n, err := im.UnreadDirectCount(ctx, uid); err == nil {
			unread = n
		}
	}
	writeJSONResponse(w, map[string]interface{}{
		"enabled":         im.IMEnabled(ctx),
		"unread":          unread,
		"companyId":       companyID,
		"companyFormHref": im.CompanySettingsFormHref(ctx, companyID),
	})
}

// SwcDirectUsersHandler GET /web/swc/direct/users?q= — internal user search for Messages tab.
func SwcDirectUsersHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !im.IMEnabled(r.Context()) {
		http.Error(w, "messages disabled", http.StatusForbidden)
		return
	}
	uid := AuthenticatedUserID(r)
	rows, err := im.SearchInternalUsers(r.Context(), uid, r.URL.Query().Get("q"), 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	type userItem struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}
	out := make([]userItem, 0, len(rows))
	for _, row := range rows {
		id, _ := orm.CoerceInt64(row["id"])
		out = append(out, userItem{
			ID:    int(id),
			Name:  strings.TrimSpace(orm.AsString(row["name"])),
			Login: strings.TrimSpace(orm.AsString(row["login"])),
		})
	}
	writeJSONResponse(w, map[string]interface{}{"users": out})
}

func SwcDirectConversationsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !im.IMEnabled(r.Context()) {
		http.Error(w, "messages disabled", http.StatusForbidden)
		return
	}
	uid := AuthenticatedUserID(r)
	items, err := im.ListDirectConversations(r.Context(), uid, 30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSONResponse(w, map[string]interface{}{"conversations": items})
}

func SwcDirectThreadHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !im.IMEnabled(r.Context()) {
		http.Error(w, "messages disabled", http.StatusForbidden)
		return
	}
	peerID, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("userId")))
	uid := AuthenticatedUserID(r)
	msgs, err := im.ListDirectThread(r.Context(), uid, peerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSONResponse(w, map[string]interface{}{"messages": msgs})
}

func SwcDirectPostHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	if !im.IMEnabled(r.Context()) {
		http.Error(w, "messages disabled", http.StatusForbidden)
		return
	}
	if !ParsePostForm(w, r) {
		return
	}
	recipientID, _ := strconv.Atoi(strings.TrimSpace(r.PostFormValue("recipient_id")))
	body := strings.TrimSpace(r.PostFormValue("body"))
	resModel := strings.TrimSpace(r.PostFormValue("res_model"))
	resID, _ := strconv.Atoi(strings.TrimSpace(r.PostFormValue("res_id")))
	uid := AuthenticatedUserID(r)
	id, err := im.PostDirectMessageLinked(r.Context(), uid, recipientID, body, resModel, resID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSONResponse(w, map[string]interface{}{"id": id})
}

func SwcDirectUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginMultipartPost(w, r, maxImportBodyBytes) {
		return
	}
	if !im.IMEnabled(r.Context()) {
		http.Error(w, "messages disabled", http.StatusForbidden)
		return
	}
	msgID, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("message_id")))
	if 0 >= msgID {
		http.Error(w, "message_id required", http.StatusBadRequest)
		return
	}
	uid := AuthenticatedUserID(r)
	ctx := r.Context()
	if err := orm.CheckModelAccess(ctx, uid, im.MessageModel, "read"); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	msg, err := orm.SearchOne(ctx, im.MessageModel, map[string]interface{}{"id": msgID})
	if err != nil {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}
	sid, _ := orm.CoerceInt64(msg["sender_id"])
	rid, _ := orm.CoerceInt64(msg["recipient_id"])
	if int(sid) != uid && int(rid) != uid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	upload, header, err := r.FormFile(importFileField)
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer upload.Close()
	content, err := io.ReadAll(upload)
	if err != nil || 0 == len(content) {
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(header.Filename)
	if name == "" {
		name = "upload"
	}
	name = filepath.Base(name)
	attID, err := orm.CreateBinaryAttachment(ctx, orm.CreateBinaryAttachmentInput{
		Name:     name,
		ResModel: im.MessageModel,
		ResID:    msgID,
		Data:     content,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   attID,
		"name": name,
		"url":  "/web/content/" + strconv.Itoa(attID),
	})
}

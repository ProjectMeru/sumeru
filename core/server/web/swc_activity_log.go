package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

const swcActivityLogRoute = "/web/swc/activity-log"

func registerSwcActivityLogRoute() {
	registerSession(http.MethodGet, swcActivityLogRoute, SwcActivityLogHandler)
}

type swcActivityLogItem struct {
	Meta       string `json:"meta"`
	Body       string `json:"body"`
	Action     string `json:"action,omitempty"`
	CreateDate string `json:"createDate,omitempty"`
}

type swcActivityLogPayload struct {
	Items   []swcActivityLogItem `json:"items"`
	HasLog  bool                 `json:"hasLog"`
	HasMore bool                 `json:"hasMore,omitempty"`
	Offset  int                  `json:"offset,omitempty"`
}

// SwcActivityLogHandler GET /web/swc/activity-log?model=&id= — sys.audit rows for the Log tab only (not mail.message).
func SwcActivityLogHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	model := strings.TrimSpace(r.URL.Query().Get("model"))
	idRaw := strings.TrimSpace(r.URL.Query().Get("id"))
	recordID, _ := strconv.ParseInt(idRaw, 10, 64)
	if model == "" || 0 >= recordID {
		http.Error(w, "model and id required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	uid := AuthenticatedUserID(r)
	if err := orm.CheckModelAccess(ctx, uid, model, "read"); err != nil {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}
	if !orm.RecordAuditLogAvailable(ctx) {
		http.Error(w, "audit not available", http.StatusNotFound)
		return
	}
	if _, err := orm.SearchOne(ctx, model, map[string]interface{}{"id": int(recordID)}); err != nil {
		http.NotFound(w, r)
		return
	}
	offset, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("offset")))
	if offset < 0 {
		offset = 0
	}
	const pageSize = 40
	rows, err := orm.SearchRecordAuditLog(ctx, model, int(recordID), pageSize, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	out := swcActivityLogPayload{
		HasLog:  true,
		Offset:  offset,
		HasMore: len(rows) >= pageSize,
		Items:   make([]swcActivityLogItem, 0, len(rows)),
	}
	for _, row := range rows {
		action := strings.TrimSpace(orm.AsString(row["action"]))
		detail := strings.TrimSpace(orm.AsString(row["detail"]))
		resModel := strings.TrimSpace(orm.AsString(row["model"]))
		if resModel == "" {
			resModel = model
		}
		body := orm.FormatAuditLogBody(ctx, resModel, action,
			orm.AsString(row["before_json"]), orm.AsString(row["after_json"]), detail)
		author := "System"
		if userID, ok := orm.CoerceInt64(row["user_id"]); ok && userID > 0 {
			if u, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": int(userID)}); err == nil {
				if nm := strings.TrimSpace(orm.AsString(u["name"])); nm != "" {
					author = nm
				} else if lg := strings.TrimSpace(orm.AsString(u["login"])); lg != "" {
					author = lg
				}
			}
		}
		when := strings.TrimSpace(orm.AsString(row["create_date"]))
		meta := author
		if when != "" {
			meta = author + " · " + when
		}
		out.Items = append(out.Items, swcActivityLogItem{
			Meta:       meta,
			Body:       body,
			Action:     action,
			CreateDate: when,
		})
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(out)
}

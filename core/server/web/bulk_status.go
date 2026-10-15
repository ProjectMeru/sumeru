package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/report"
)

// BulkStatusHandler GET /web/bulk/status?id= — job state for queued import/export batches.
func BulkStatusHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	id, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	if 0 >= id {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	uid := AuthenticatedUserID(r)
	batch, err := orm.SearchOne(r.Context(), report.BulkModelName, map[string]interface{}{"id": id})
	if err != nil || 0 == len(batch) {
		http.NotFound(w, r)
		return
	}
	ownerID, _ := orm.CoerceInt64(batch["user_id"])
	if int(ownerID) != uid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	state := strings.TrimSpace(orm.AsString(batch["state"]))
	summary := strings.TrimSpace(orm.AsString(batch["result_summary"]))
	attID, _ := orm.CoerceInt64(batch["attachment_id"])
	out := map[string]interface{}{
		"id":      id,
		"state":   state,
		"summary": summary,
	}
	if attID > 0 {
		out["attachmentUrl"] = contentRoutePrefix + strconv.FormatInt(attID, 10)
	}
	writeJSONResponse(w, out)
}

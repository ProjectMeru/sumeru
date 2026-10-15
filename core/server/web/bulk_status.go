package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/report"
)

func bulkBatchOwnedBy(batch map[string]interface{}, uid int) bool {
	ownerID, _ := orm.CoerceInt64(batch["user_id"])
	return uid > 0 && int(ownerID) == uid
}

// loadOwnedBulkBatch loads sys.bulk.import when batchID belongs to the logged-in user.
func loadOwnedBulkBatch(w http.ResponseWriter, r *http.Request, batchID int) (map[string]interface{}, bool) {
	if 0 >= batchID {
		http.Error(w, "id required", http.StatusBadRequest)
		return nil, false
	}
	batch, err := orm.SearchOne(r.Context(), report.BulkModelName, map[string]interface{}{"id": batchID})
	if err != nil || 0 == len(batch) {
		http.NotFound(w, r)
		return nil, false
	}
	if !bulkBatchOwnedBy(batch, AuthenticatedUserID(r)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, false
	}
	return batch, true
}

// BulkStatusHandler GET /web/bulk/status?id= — job state for queued import/export batches.
func BulkStatusHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	id, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	batch, ok := loadOwnedBulkBatch(w, r, id)
	if !ok {
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

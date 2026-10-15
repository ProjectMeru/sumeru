package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/report"
)

func registerSwcImportRoute() {
	registerSession(http.MethodGet, swcImportRoute, SwcImportHandler)
}

// SwcImportHandler GET /web/swc/import?batch= — import wizard bootstrap JSON.
func SwcImportHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	batchID, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("batch")))
	if 0 >= batchID {
		http.Error(w, "batch required", http.StatusBadRequest)
		return
	}
	batch, err := orm.SearchOne(r.Context(), report.BulkModelName, map[string]interface{}{"id": batchID})
	if err != nil || 0 == len(batch) {
		http.NotFound(w, r)
		return
	}
	targetModel := orm.AsString(batch["target_model"])
	if !requireModelAccess(w, r, targetModel, "create") {
		return
	}
	var headers []string
	_ = json.Unmarshal([]byte(orm.AsString(batch["csv_headers"])), &headers)
	mapping := map[string]string{}
	_ = json.Unmarshal([]byte(orm.AsString(batch["column_mapping"])), &mapping)
	var fields []string
	_ = json.Unmarshal([]byte(orm.AsString(batch["selected_fields"])), &fields)
	modelInst, _ := orm.Registry[targetModel]
	fieldLabels := map[string]string{}
	if modelInst != nil {
		for _, f := range modelInst.Fields() {
			if f.Name != "" {
				fieldLabels[f.Name] = f.String
			}
		}
	}
	writeJSONResponse(w, map[string]interface{}{
		"batchId":     batchID,
		"targetModel": targetModel,
		"importMode":  orm.AsString(batch["import_mode"]),
		"headers":     headers,
		"mapping":     mapping,
		"fields":      fields,
		"fieldLabels": fieldLabels,
		"state":       orm.AsString(batch["state"]),
		"nextUrl":     orm.AsString(batch["next_url"]),
	})
}

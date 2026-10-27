package web

import (
	"net/http"
	"strings"

	"sumeru/core/engine/swcmeta"
	"sumeru/core/orm"
	"sumeru/core/report"
)

func requireExportEntitlement(w http.ResponseWriter, r *http.Request, modelName string, fields []string, format string) bool {
	ctx := r.Context()
	uid := orm.SecurityUID(ctx)
	if err := orm.CheckFieldReadAccess(ctx, uid, modelName, fields); err != nil {
		http.Error(w, forbiddenMessage, http.StatusForbidden)
		return false
	}
	actionID := report.ParseActionIDParam(r.URL.Query().Get(actionIDField))
	if actionID <= 0 {
		return true
	}
	row, err := loadWindowAction(ctx, actionID)
	if err != nil {
		http.Error(w, forbiddenMessage, http.StatusForbidden)
		return false
	}
	resModel := strings.TrimSpace(orm.AsString(row["res_model"]))
	if resModel != "" && resModel != modelName {
		http.Error(w, forbiddenMessage, http.StatusForbidden)
		return false
	}
	modes := splitViewModes(strings.TrimSpace(orm.AsString(row["view_mode"])))
	caps := swcmeta.ExportEntitlementsForModel(ctx, modelName, modes)
	if !caps.HasDownload() {
		http.Error(w, forbiddenMessage, http.StatusForbidden)
		return false
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		return true
	}
	for _, allowed := range caps.DownloadFormats {
		if allowed == format {
			return true
		}
	}
	http.Error(w, forbiddenMessage, http.StatusForbidden)
	return false
}

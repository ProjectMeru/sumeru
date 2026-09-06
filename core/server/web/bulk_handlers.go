package web

import (
	"io"
	"net/http"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/report"
)

// BulkUploadHandler POST /web/bulk/upload — stage CSV and open mapping view.
func BulkUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginMultipartPost(w, r, maxImportBodyBytes) {
		return
	}
	modelName := strings.TrimSpace(r.FormValue(importModelField))
	if modelName == "" {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}
	if !requireModelAccess(w, r, modelName, "create") {
		return
	}
	upload, _, err := r.FormFile(importFileField)
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer upload.Close()
	content, err := io.ReadAll(upload)
	if err != nil || len(content) == 0 {
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}
	fields := report.ParseFieldsParam(r.FormValue(reportFieldsParam))
	mode := strings.TrimSpace(r.FormValue(importModeField))
	if mode == "" {
		mode = report.ImportModeCreate
	}
	uid := AuthenticatedUserID(r)
	actionID := report.ParseActionIDParam(r.FormValue(actionIDField))
	batchID, err := report.CreateBatch(r.Context(), report.CreateBatchInput{
		TargetModel:    modelName,
		ImportMode:     mode,
		SelectedFields: fields,
		CSVContent:     content,
		NextURL:        SafeWebNext(r.FormValue(nextField), homeRoute),
		UserID:         uid,
		ActionID:       actionID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	actionID = report.BulkImportFormActionID(r.Context())
	http.Redirect(w, r, report.MappingFormURL(batchID, actionID), http.StatusSeeOther)
}

// BulkConfirmHandler POST /web/bulk/confirm — run import after mapping (non-object-action path).
func BulkConfirmHandler(w http.ResponseWriter, r *http.Request) {
	if !RequirePOST(w, r) {
		return
	}
	if !validateSessionCSRF(w, r) || !requireLogin(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	batchID := report.ParseActionIDParam(r.FormValue("id"))
	if batchID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	mapping, err := report.ParseMappingJSON(r.FormValue("column_mapping"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	skipInvalid := r.FormValue("skip_invalid") == "1"
	result, err := report.ExecuteBulkImport(r.Context(), report.ExecuteBulkImportInput{
		BatchID:     batchID,
		Mapping:     mapping,
		SkipInvalid: skipInvalid,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	batch, _ := orm.SearchOne(r.Context(), report.BulkModelName, map[string]interface{}{"id": batchID})
	next := orm.AsString(batch["next_url"])
	redirectWithWebMessage(w, r, next, report.ImportFlashMessage(result))
}

// BulkCancelHandler POST /web/bulk/cancel
func BulkCancelHandler(w http.ResponseWriter, r *http.Request) {
	if !RequirePOST(w, r) {
		return
	}
	if !validateSessionCSRF(w, r) || !requireLogin(w, r) {
		return
	}
	_ = r.ParseForm()
	batchID := report.ParseActionIDParam(r.FormValue("id"))
	if batchID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	if err := report.CancelBatch(r.Context(), batchID); err != nil {
		http.Error(w, "cancel failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	batch, err := orm.SearchOne(r.Context(), report.BulkModelName, map[string]interface{}{"id": batchID})
	next := homeRoute
	if err == nil {
		next = SafeWebNext(orm.AsString(batch["next_url"]), homeRoute)
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

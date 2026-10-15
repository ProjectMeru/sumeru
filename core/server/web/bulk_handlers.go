package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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
	upload, fileHeader, err := r.FormFile(importFileField)
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
	filename := ""
	if fileHeader != nil {
		filename = fileHeader.Filename
	}
	content, err = report.ParseUploadContent(filename, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := orm.ValidateAttachmentMIME("", content); err != nil {
		http.Error(w, "file type not allowed", http.StatusBadRequest)
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
	http.Redirect(w, r, importWizardURL(batchID), http.StatusSeeOther)
}

func importWizardURL(batchID int) string {
	return "/web/import?batch=" + strconv.Itoa(batchID)
}

// BulkPreviewHandler POST JSON preview for a staged batch.
func BulkPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginJSONPost(w, r) {
		return
	}
	var body struct {
		BatchID int               `json:"batch_id"`
		Mapping map[string]string `json:"column_mapping"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.BatchID <= 0 {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	out, err := report.PreviewBulkImport(r.Context(), report.PreviewBulkImportInput{
		BatchID: body.BatchID,
		Mapping: body.Mapping,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSONResponse(w, out)
}

// BulkSaveImportTemplateHandler POST JSON — persist column mapping as sys.import.template.
func BulkSaveImportTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginJSONPost(w, r) {
		return
	}
	var body struct {
		Name    string            `json:"name"`
		BatchID int               `json:"batch_id"`
		Mapping map[string]string `json:"column_mapping"`
		Shared  bool              `json:"shared"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.BatchID <= 0 {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	batch, err := orm.SearchOne(r.Context(), report.BulkModelName, map[string]interface{}{"id": body.BatchID})
	if err != nil || 0 == len(batch) {
		http.NotFound(w, r)
		return
	}
	targetModel := orm.AsString(batch["target_model"])
	if !requireModelAccess(w, r, targetModel, "create") {
		return
	}
	mapping := body.Mapping
	if 0 == len(mapping) {
		mapping, err = report.ParseMappingJSON(orm.AsString(batch["column_mapping"]))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		http.Error(w, "mapping encode failed", http.StatusBadRequest)
		return
	}
	selected := orm.AsString(batch["selected_fields"])
	importMode := orm.AsString(batch["import_mode"])
	uid := AuthenticatedUserID(r)
	templateID, err := report.SaveImportTemplate(r.Context(), name, targetModel, importMode, selected, string(mappingJSON), uid, body.Shared)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSONResponse(w, map[string]interface{}{"template_id": templateID})
}

// BulkDryRunHandler POST JSON dry-run validation for entire file.
func BulkDryRunHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginJSONPost(w, r) {
		return
	}
	var body struct {
		BatchID int               `json:"batch_id"`
		Mapping map[string]string `json:"column_mapping"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.BatchID <= 0 {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	preview, messages, err := report.DryRunBulkImport(r.Context(), report.PreviewBulkImportInput{
		BatchID: body.BatchID,
		Mapping: body.Mapping,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSONResponse(w, map[string]interface{}{"preview": preview, "messages": messages})
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
	if 0 >= batchID {
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
	if 0 >= batchID {
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

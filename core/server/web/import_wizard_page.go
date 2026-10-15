package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/render"
)

// ImportWizardPageHandler GET /web/import?batch= — shell page for SWC import wizard.
func ImportWizardPageHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	batchIDStr := strings.TrimSpace(r.URL.Query().Get("batch"))
	if batchIDStr == "" {
		http.Error(w, "batch required", http.StatusBadRequest)
		return
	}
	batchID, err := strconv.Atoi(batchIDStr)
	if err != nil || 0 >= batchID {
		http.Error(w, "invalid batch", http.StatusBadRequest)
		return
	}
	if _, ok := loadOwnedBulkBatch(w, r, batchID); !ok {
		return
	}
	renderShellPage(w, r, shellPageOpts{
		Route:         importWizardRoute,
		InnerTemplate: render.TemplatePagesImportWizard,
		InnerData: map[string]interface{}{
			"BatchID":   batchIDStr,
			"CSRFToken": CSRFTokenForRequest(r),
		},
		Page: render.PageData{Title: "Import data"},
	})
}

package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/report"
)

const (
	reportPrintRoute    = "/web/report/print"
	exportPivotRoute    = "/web/export/pivot"
	exportGraphRoute    = "/web/export/graph"
)

func registerReportRoutes() {
	registerSession(http.MethodGet, reportPrintRoute, ReportPrintHandler)
	registerSession(http.MethodGet, exportPivotRoute, ExportPivotHandler)
	registerSession(http.MethodGet, exportGraphRoute, ExportGraphHandler)
}

// ReportPrintHandler GET /web/report/print?report_id=&id=
func ReportPrintHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	reportID, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("report_id")))
	recordID, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	data, filename, err := report.RenderReportActionPDF(r.Context(), reportID, recordID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	_, _ = w.Write(data)
}

// ExportPivotHandler GET /web/export/pivot
func ExportPivotHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	modelName := strings.TrimSpace(r.URL.Query().Get(importModelField))
	if modelName == "" {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}
	if !requireModelAccess(w, r, modelName, "read") {
		return
	}
	groupFields := report.ParseFieldsParam(r.URL.Query().Get("group_by"))
	measures := report.ParseFieldsParam(r.URL.Query().Get("measures"))
	if len(groupFields) == 0 || len(measures) == 0 {
		http.Error(w, "group_by and measures required", http.StatusBadRequest)
		return
	}
	domain := analyticsExportDomain(r.Context(), r, modelName)
	data, err := report.ExportPivotCSV(r.Context(), report.PivotExportInput{
		Model:       modelName,
		Domain:      domain,
		GroupFields: groupFields,
		Measures:    measures,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+report.ExportFilename(modelName+"_pivot", "csv"))
	_, _ = w.Write(data)
}

// ExportGraphHandler GET /web/export/graph
func ExportGraphHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	modelName := strings.TrimSpace(r.URL.Query().Get(importModelField))
	if modelName == "" {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}
	if !requireModelAccess(w, r, modelName, "read") {
		return
	}
	groupField := strings.TrimSpace(r.URL.Query().Get("group_by"))
	measureField := strings.TrimSpace(r.URL.Query().Get("measure"))
	if groupField == "" || measureField == "" {
		http.Error(w, "group_by and measure required", http.StatusBadRequest)
		return
	}
	domain := analyticsExportDomain(r.Context(), r, modelName)
	data, err := report.ExportGraphCSV(r.Context(), report.GraphExportInput{
		Model:        modelName,
		Domain:       domain,
		GroupField:   groupField,
		MeasureField: measureField,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+report.ExportFilename(modelName+"_graph", "csv"))
	_, _ = w.Write(data)
}

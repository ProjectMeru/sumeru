package web

import (
	"net/http"
	"strconv"
	"strings"
)

// ImportCSVHandler imports CSV rows — redirects to bulk upload staging flow.
func ImportCSVHandler(w http.ResponseWriter, r *http.Request) {
	BulkUploadHandler(w, r)
}

func importCSVFlashMessage(createdCount int) string {
	return "imported_" + strconv.Itoa(createdCount)
}

func importableRowValues(header, record []string, allowedFields map[string]struct{}) map[string]interface{} {
	values := map[string]interface{}{}
	for columnIndex, columnName := range header {
		if !isImportableColumn(columnName, allowedFields) {
			continue
		}
		if columnIndex >= len(record) {
			continue
		}
		values[columnName] = coerceCSVValue(record[columnIndex])
	}
	return values
}

func isImportableColumn(columnName string, allowedFields map[string]struct{}) bool {
	columnName = strings.TrimSpace(columnName)
	if columnName == "" || columnName == workspaceRecordIDParam {
		return false
	}
	_, allowed := allowedFields[columnName]
	return allowed
}

func coerceCSVValue(raw string) interface{} {
	value := strings.TrimSpace(raw)
	if value == "" {
		return value
	}
	switch strings.ToLower(value) {
	case "true":
		return true
	case "false":
		return false
	}
	if number, err := strconv.ParseInt(value, 10, 64); err == nil {
		return number
	}
	return value
}

package web

import (
	"net/http"
	"strings"

	"sumeru/core/module"
	"sumeru/core/orm"
)

// moduleRow holds normalized fields from a sys.module ORM row.
type moduleRow struct {
	Name        string
	DisplayName string
	Author      string
	Version     string
	Description string
	State       string
	Application bool
	Active      bool
}

func moduleDisplayName(name, displayName string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName != "" {
		return displayName
	}
	return name
}

// parseModuleRow extracts common module fields from an ORM row map.
func parseModuleRow(row map[string]interface{}) (moduleRow, bool) {
	name := orm.AsString(row["name"])
	if name == "" {
		return moduleRow{}, false
	}
	displayName := orm.AsString(row["display_name"])
	return moduleRow{
		Name:        name,
		DisplayName: moduleDisplayName(name, displayName),
		Author:      orm.AsString(row["author"]),
		Version:     orm.AsString(row["version"]),
		Description: orm.AsString(row["description"]),
		State:       orm.AsString(row["state"]),
		Application: orm.AsBool(row["application"]),
		Active:      orm.AsBool(row["active"]),
	}, true
}

// listModulesOr500 loads all modules or writes a 500 response and returns ok=false.
func listModulesOr500(w http.ResponseWriter, r *http.Request, logMessage string) ([]map[string]interface{}, bool) {
	moduleRows, err := module.ListModules(r.Context())
	if err != nil {
		WebLogEvent(r.Context(), WebLogInput{
			Route: r.URL.Path, Message: logMessage,
			Operation: "load", Status: "failure", Err: err,
		})
		http.Error(w, "Failed to list modules", http.StatusInternalServerError)
		return nil, false
	}
	return moduleRows, true
}

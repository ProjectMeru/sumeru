package orm

import (
	"sort"
	"strings"
)

// modelExtendedByModule lists models whose fields were extended by each module.
var modelExtendedByModule = make(map[string][]string)

// RecordModelExtendedBy records that moduleName added fields to modelName.
func RecordModelExtendedBy(modelName, moduleName string) {
	modelName = strings.TrimSpace(modelName)
	moduleName = strings.TrimSpace(moduleName)
	if modelName == "" || moduleName == "" {
		return
	}
	for _, existing := range modelExtendedByModule[moduleName] {
		if existing == modelName {
			return
		}
	}
	modelExtendedByModule[moduleName] = append(modelExtendedByModule[moduleName], modelName)
}

// ModelsExtendedByModule lists model names that moduleName extended via inherit= structs.
func ModelsExtendedByModule(moduleName string) []string {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return nil
	}
	names := append([]string(nil), modelExtendedByModule[moduleName]...)
	sort.Strings(names)
	return names
}

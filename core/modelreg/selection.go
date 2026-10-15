package modelreg

import (
	"reflect"
	"strings"
)

var selectionByType = map[string][][]string{}

// RegisterSelectionForType records selection options for a Go type.
func RegisterSelectionForType(rt reflect.Type, options [][]string) {
	if rt == nil || 0 == len(options) {
		return
	}
	selectionByType[rt.String()] = options
	if name := rt.Name(); name != "" {
		selectionByType[name] = options
	}
}

func selectionForFieldType(fieldType reflect.Type) [][]string {
	key := selectionTypeKey(fieldType)
	if key == "" {
		return nil
	}
	if opts, ok := selectionByType[key]; ok {
		return opts
	}
	if i := strings.LastIndex(key, "."); i >= 0 {
		if opts, ok := selectionByType[key[i+1:]]; ok {
			return opts
		}
	}
	return nil
}

func selectionTypeKey(fieldType reflect.Type) string {
	if fieldType.Kind() == reflect.Struct && markerBaseName(fieldType) == "Selection" {
		arg := genericArgString(fieldType)
		if arg == "" {
			return ""
		}
		return arg
	}
	return ""
}

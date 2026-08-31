package modelmeta

import "reflect"

// ModelNameFromStruct reads the technical model name from an embedded ModelMeta tag,
// or falls back to ModelNameFromGo for the struct name.
func ModelNameFromStruct(st reflect.Type) (string, error) {
	spec, err := ModelSpecFromStruct(st)
	if err != nil {
		return "", err
	}
	return spec.Name, nil
}

package modelmeta

import "reflect"

// IsEmbeddedModelMeta reports whether f is an anonymous embedded ModelMeta field.
func IsEmbeddedModelMeta(field reflect.StructField) bool {
	if !field.Anonymous {
		return false
	}
	t := field.Type
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t == reflect.TypeOf(ModelMeta{})
}

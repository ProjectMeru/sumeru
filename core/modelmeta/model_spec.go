package modelmeta

import (
	"fmt"
	"reflect"
)

// ModelSpec describes how a Go struct maps to an ORM model name.
type ModelSpec struct {
	Name               string
	Extend             bool   // true when the struct uses inherit= to extend an existing model
	DelegationParent   string // set when inherits= names a parent model (_inherits delegation)
}

// ModelSpecFromStruct reads model= or inherit= from an embedded ModelMeta tag.
func ModelSpecFromStruct(structType reflect.Type) (ModelSpec, error) {
	for structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return ModelSpec{}, nil
	}
	tags, ok, err := embeddedModelTags(structType)
	if err != nil {
		return ModelSpec{}, err
	}
	if !ok {
		return ModelSpec{Name: ModelNameFromGo(structType.Name()), Extend: false}, nil
	}
	return ModelSpecFromTags(tags, structType.Name())
}

// ModelSpecFromTags builds a ModelSpec from parsed embedded ModelMeta tags and the Go type name.
func ModelSpecFromTags(tags FieldTags, goName string) (ModelSpec, error) {
	if err := validateModelInheritExclusive(tags); err != nil {
		return ModelSpec{}, fmt.Errorf("struct %s: %w", goName, err)
	}
	if tags.Inherit != "" {
		return ModelSpec{Name: tags.Inherit, Extend: true}, nil
	}
	if tags.Inherits != "" {
		return ModelSpec{Name: ModelNameFromGo(goName), DelegationParent: tags.Inherits}, nil
	}
	if tags.Model == "-" {
		return ModelSpec{Name: "-", Extend: false}, nil
	}
	if tags.Model != "" {
		return ModelSpec{Name: tags.Model, Extend: false}, nil
	}
	return ModelSpec{Name: ModelNameFromGo(goName), Extend: false}, nil
}

// ModelNameFromStruct reads the technical model name from an embedded ModelMeta tag,
// or falls back to ModelNameFromGo for the struct name.
func ModelNameFromStruct(structType reflect.Type) (string, error) {
	spec, err := ModelSpecFromStruct(structType)
	if err != nil {
		return "", err
	}
	return spec.Name, nil
}

func embeddedModelTags(structType reflect.Type) (FieldTags, bool, error) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !IsEmbeddedModelMeta(field) {
			continue
		}
		tags, err := ParseFieldTag(string(field.Tag.Get("sumeru")))
		if err != nil {
			return FieldTags{}, false, err
		}
		return tags, true, nil
	}
	return FieldTags{}, false, nil
}

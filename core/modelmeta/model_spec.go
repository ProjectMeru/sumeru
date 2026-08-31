package modelmeta

import (
	"fmt"
	"reflect"
)

// ModelSpec describes how a Go struct maps to an ORM model name.
type ModelSpec struct {
	Name   string
	Extend bool // true when the struct uses inherit= to extend an existing model
}

// ModelSpecFromStruct reads model= or inherit= from an embedded ModelMeta tag.
func ModelSpecFromStruct(st reflect.Type) (ModelSpec, error) {
	for st.Kind() == reflect.Pointer {
		st = st.Elem()
	}
	if st.Kind() != reflect.Struct {
		return ModelSpec{}, nil
	}
	tags, ok, err := embeddedModelTags(st)
	if err != nil {
		return ModelSpec{}, err
	}
	if !ok {
		return ModelSpec{Name: ModelNameFromGo(st.Name()), Extend: false}, nil
	}
	if tags.Model != "" && tags.Inherit != "" {
		return ModelSpec{}, fmt.Errorf("struct %s: model= and inherit= are mutually exclusive", st.Name())
	}
	if tags.Inherit != "" {
		return ModelSpec{Name: tags.Inherit, Extend: true}, nil
	}
	if tags.Model == "-" {
		return ModelSpec{Name: "-", Extend: false}, nil
	}
	if tags.Model != "" {
		return ModelSpec{Name: tags.Model, Extend: false}, nil
	}
	return ModelSpec{Name: ModelNameFromGo(st.Name()), Extend: false}, nil
}

func embeddedModelTags(st reflect.Type) (FieldTags, bool, error) {
	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		if !f.Anonymous {
			continue
		}
		t := f.Type
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t != reflect.TypeOf(ModelMeta{}) {
			continue
		}
		tags, err := parseSumeruTag(string(f.Tag.Get("sumeru")))
		if err != nil {
			return FieldTags{}, false, err
		}
		return tags, true, nil
	}
	return FieldTags{}, false, nil
}

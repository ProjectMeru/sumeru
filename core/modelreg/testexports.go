package modelreg

import (
	"reflect"

	"sumeru/core/modelmeta"
	"sumeru/core/orm"
)

func ParseSelectionForTest(raw string) [][]string { return parseSelection(raw) }

func ParseDefaultForTest(raw string, fieldType orm.FieldType) interface{} {
	return parseDefault(raw, fieldType)
}

func MapMarkerTypeForTest(marker string) (orm.FieldType, string, error) {
	return mapMarkerType(marker)
}

type RegisterCtxForTest struct {
	inner *registerCtx
}

func NewRegisterCtxForTest() *RegisterCtxForTest {
	return &RegisterCtxForTest{inner: &registerCtx{
		typeNames: map[string]string{},
		byName:    map[string]reflect.Type{},
	}}
}

func (c *RegisterCtxForTest) SetTypeMapping(typeKey, modelName, goName string, typ reflect.Type) *RegisterCtxForTest {
	c.inner.typeNames[typeKey] = modelName
	c.inner.byName[goName] = typ
	return c
}

func (c *RegisterCtxForTest) ResolveComodel(fieldType reflect.Type, tags modelmeta.FieldTags) (string, error) {
	return c.inner.resolveComodel(fieldType, tags)
}

func SelectionOptionsFromPackageForTest(pkgDir, typeName string) [][]string {
	return selectionOptionsFromPackage(pkgDir, typeName)
}

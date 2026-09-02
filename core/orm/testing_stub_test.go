package orm

import "testing"

// stubModel implements Model for in-package tests. modelreg cannot be used here
// because it imports orm (import cycle).
type stubModel struct {
	name   string
	fields []FieldDefinition
}

func (m stubModel) ModelName() string         { return m.name }
func (m stubModel) Fields() []FieldDefinition { return m.fields }

func registerStubModel(t *testing.T, name string, fields []FieldDefinition) stubModel {
	t.Helper()
	m := stubModel{name: name, fields: fields}
	RegisterModelWithModule(m, "test")
	t.Cleanup(func() { delete(Registry, name) })
	return m
}

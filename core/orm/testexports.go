package orm

import "testing"

type StubModelForTest struct {
	name   string
	fields []FieldDefinition
}

func (m StubModelForTest) ModelName() string         { return m.name }
func (m StubModelForTest) Fields() []FieldDefinition { return m.fields }

func NewStubModelForTest(name string, fields []FieldDefinition) StubModelForTest {
	return StubModelForTest{name: name, fields: fields}
}

// RegisterStubModelForTest registers a minimal model for external tests.
func RegisterStubModelForTest(t *testing.T, name string, fields []FieldDefinition) StubModelForTest {
	t.Helper()
	m := StubModelForTest{name: name, fields: fields}
	RegisterModelWithModule(m, "test")
	t.Cleanup(func() { delete(Registry, name) })
	return m
}

func BuildSearchWhereClauseForTest(modelName string, domain [][]interface{}) (string, []interface{}, error) {
	return buildSearchWhereClause(modelName, domain)
}

func CoerceFieldValueForTest(fieldDef FieldDefinition, v interface{}) (interface{}, error) {
	return coerceFieldValue(fieldDef, v)
}

func SplitDomainORPrefixForTest(domain [][]interface{}) (orCount int, leaves [][]interface{}, ok bool) {
	return splitDomainORPrefix(domain)
}

func BuildAndWhereClausesForTest(modelName string, groups [][][]interface{}) (string, []interface{}, error) {
	return buildAndWhereClauses(modelName, groups)
}

func RecordMatchesDomainForTest(rec map[string]interface{}, domain [][]interface{}) bool {
	return RecordMatchesDomain(rec, domain)
}

// SetDBForTest replaces the global DB handle for external tests. Call ResetDBForTest in t.Cleanup.
func SetDBForTest(w DBWrapper) {
	DB = w
}

// ResetDBForTest clears the global DB handle.
func ResetDBForTest() {
	DB = nil
	readDB = nil
	readReplicaReady = false
}

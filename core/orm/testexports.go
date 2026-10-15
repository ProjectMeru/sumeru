package orm

import (
	"context"
	"testing"
)

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
// StampActiveCompanyOnCreateForTest applies create-time company_id defaulting for tests.
func StampActiveCompanyOnCreateForTest(ctx context.Context, modelName string, prepared map[string]interface{}) {
	stampActiveCompanyOnCreate(ctx, modelName, prepared)
}

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

// FormatAuditLogBodyForTest exposes audit log formatting for tests.
func FormatAuditLogBodyForTest(ctx context.Context, model, action, beforeJSON, afterJSON, detail string) string {
	return FormatAuditLogBody(ctx, model, action, beforeJSON, afterJSON, detail)
}

func RecordAuditLogAvailableForTest(ctx context.Context) bool {
	return RecordAuditLogAvailable(ctx)
}

func SearchRecordAuditLogForTest(ctx context.Context, model string, resID int, limit, offset int) ([]map[string]interface{}, error) {
	return SearchRecordAuditLog(ctx, model, resID, limit, offset)
}

func SplitDomainORPrefixForTest(domain [][]interface{}) (orCount int, leaves [][]interface{}, ok bool) {
	return splitDomainORPrefix(domain)
}

func BuildAndWhereClausesForTest(modelName string, groups [][][]interface{}) (string, []interface{}, error) {
	return buildAndWhereClauses(modelName, groups)
}

func RecordMatchesDomainForTest(modelName string, rec map[string]interface{}, domain [][]interface{}) bool {
	return RecordMatchesDomain(modelName, rec, domain)
}

func ResolveActionRecordForTest(ctx context.Context, actionID int, actionQuery string) (modelName string, coreID int, err error) {
	return ResolveActionRecord(ctx, actionID, actionQuery)
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

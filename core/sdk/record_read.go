package sdk

import (
	"context"

	"sumeru/core/orm"
)

// SearchOne finds one row by exact equality on all map keys.
func SearchOne(ctx context.Context, modelName string, criteria map[string]interface{}) (map[string]interface{}, error) {
	return orm.SearchOne(ctx, modelName, criteria)
}

// Search runs a conjunctive domain and returns full rows.
func Search(ctx context.Context, modelName string, domain [][]interface{}) ([]map[string]interface{}, error) {
	return orm.Search(ctx, modelName, domain)
}

// SearchRead is Search plus a column list, limit, offset, and order.
// An empty field list keeps every column.
func SearchRead(ctx context.Context, modelName string, domain [][]interface{}, fields []string, limit, offset int, orderBy string) ([]map[string]interface{}, error) {
	return orm.SearchRead(ctx, modelName, domain, fields, limit, offset, orderBy)
}

// SearchCount returns how many rows match the domain.
func SearchCount(ctx context.Context, modelName string, domain [][]interface{}) (int, error) {
	return orm.SearchCount(ctx, modelName, domain)
}

// SearchPage returns one page of rows.
func SearchPage(ctx context.Context, modelName string, domain [][]interface{}, limit, offset int, orderBy string) ([]map[string]interface{}, error) {
	return orm.SearchPage(ctx, modelName, domain, limit, offset, orderBy)
}

// Read loads the given ids. An empty field list returns full rows.
func Read(ctx context.Context, modelName string, ids []int, fields []string) ([]map[string]interface{}, error) {
	if 0 == len(ids) {
		return []map[string]interface{}{}, nil
	}
	return SearchRead(ctx, modelName, idDomain(ids), fields, 0, 0, "")
}

// ResolveXmlId resolves module.xml_id to a database id and model name.
func ResolveXmlId(ctx context.Context, xmlid string) (int, string, error) {
	return orm.ResolveXmlId(ctx, xmlid)
}

// NameSearch returns id and display label rows for many2one completion.
func NameSearch(ctx context.Context, modelName, query string, limit int) ([]map[string]interface{}, error) {
	return orm.NameSearch(ctx, modelName, query, limit)
}

// DefaultGet builds a new-row map from field defaults, then applies defaults (action default_* values).
func DefaultGet(ctx context.Context, modelName string, fields []string, defaults map[string]interface{}) (map[string]interface{}, error) {
	return orm.DefaultGet(ctx, modelName, fields, defaults)
}

// RecordMatchesDomain reports whether one row satisfies a domain.
func RecordMatchesDomain(modelName string, record map[string]interface{}, domain [][]interface{}) bool {
	return orm.RecordMatchesDomain(modelName, record, domain)
}

// RegisterSearchResultFilter adds a post-search row filter.
func RegisterSearchResultFilter(fn func(ctx context.Context, model string, rows []map[string]interface{}) []map[string]interface{}) {
	orm.RegisterSearchResultFilter(fn)
}

// AsString reads a cell as text.
func AsString(v interface{}) string { return orm.AsString(v) }

// AsBool reads a cell as a boolean.
func AsBool(v interface{}) bool { return orm.AsBool(v) }

// CoerceInt64 reads a cell as an integer.
func CoerceInt64(v interface{}) (int64, bool) { return orm.CoerceInt64(v) }

func idDomain(ids []int) [][]interface{} {
	list := make([]interface{}, len(ids))
	for i, id := range ids {
		list[i] = id
	}
	return [][]interface{}{{"id", "in", list}}
}

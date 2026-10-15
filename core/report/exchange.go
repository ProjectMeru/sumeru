package report

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// FetchRows loads up to maxExportRows for export using an optional domain.
func FetchRows(ctx context.Context, modelName string, domain [][]interface{}, recordID int) ([]map[string]interface{}, error) {
	modelInst, ok := orm.Registry[modelName]
	if !ok {
		return nil, fmt.Errorf("unknown model %q", modelName)
	}
	uid := orm.SecurityUID(ctx)
	if recordID > 0 {
		row, err := orm.SearchOne(ctx, modelName, map[string]interface{}{"id": recordID})
		if err != nil {
			return nil, err
		}
		orm.RedactRecordForRead(ctx, uid, modelName, row)
		return []map[string]interface{}{row}, nil
	}
	rows, err := orm.SearchLimit(ctx, modelName, domain, maxExportRows)
	if err != nil {
		return nil, err
	}
	orm.RedactSearchResults(ctx, uid, modelName, rows)
	_ = modelInst
	return rows, nil
}

// FetchRowsUnlimitedForTest exposes paged export fetch for external tests.
func FetchRowsUnlimitedForTest(ctx context.Context, modelName string, domain [][]interface{}, recordID int) ([]map[string]interface{}, error) {
	return FetchRowsUnlimited(ctx, modelName, domain, recordID)
}

// FetchRowsUnlimited loads all rows matching domain (paged) for async export.
func FetchRowsUnlimited(ctx context.Context, modelName string, domain [][]interface{}, recordID int) ([]map[string]interface{}, error) {
	if recordID > 0 {
		return FetchRows(ctx, modelName, domain, recordID)
	}
	const page = 500
	var all []map[string]interface{}
	offset := 0
	for {
		chunk, err := orm.SearchPage(ctx, modelName, domain, page, offset, "id ASC")
		if err != nil {
			return nil, err
		}
		if 0 == len(chunk) {
			break
		}
		uid := orm.SecurityUID(ctx)
		orm.RedactSearchResults(ctx, uid, modelName, chunk)
		all = append(all, chunk...)
		if len(chunk) < page {
			break
		}
		offset += len(chunk)
	}
	return all, nil
}

// ValidateFields filters field names against model metadata.
func ValidateFields(modelName string, fields []string) ([]string, error) {
	modelInst, ok := orm.Registry[modelName]
	if !ok {
		return nil, fmt.Errorf("unknown model %q", modelName)
	}
	allowed := allowedFieldNames(modelInst)
	var out []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if _, ok := allowed[f]; !ok {
			return nil, fmt.Errorf("field %q not allowed on %s", f, modelName)
		}
		out = append(out, f)
	}
	if 0 == len(out) {
		return nil, fmt.Errorf("no fields selected")
	}
	return out, nil
}

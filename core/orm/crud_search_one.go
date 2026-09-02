package orm

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// CriteriaToDomain converts a SearchOne equality map into a conjunctive domain
// (sorted keys so SQL placeholder order is stable).
func CriteriaToDomain(criteria map[string]interface{}) [][]interface{} {
	if len(criteria) == 0 {
		return nil
	}
	keys := make([]string, 0, len(criteria))
	for k := range criteria {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	domain := make([][]interface{}, 0, len(keys))
	for _, k := range keys {
		domain = append(domain, []interface{}{k, "=", criteria[k]})
	}
	return domain
}

// SearchOne finds a single record matching the criteria.
// Record rules are compiled into the WHERE clause (same SQL as Search), not checked after fetch.
func SearchOne(ctx context.Context, modelName string, criteria map[string]interface{}) (result map[string]interface{}, err error) {
	start := time.Now()
	defer func() {
		logORMOperationKV(ctx, start, "search_one", modelName, err, "has_row", result != nil)
	}()
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not registered", modelName)
	}
	uid := SecurityUID(ctx)
	if !SecurityBypass(ctx) {
		if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
			return nil, err
		}
	}

	domain := CriteriaToDomain(criteria)
	if len(domain) == 0 {
		return nil, fmt.Errorf("search criteria required")
	}

	whereClause, args, err := BuildWhereWithRecordRules(ctx, uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	table, err := QuotedTableForModel(modelName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", table, whereClause)
	rows, err := DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		err = sql.ErrNoRows
		return nil, err
	}

	cols, _ := rows.Columns()
	result, err = scanRowToMap(cols, rows)
	if err != nil {
		return nil, err
	}
	enrichRecordForRead(ctx, uid, modelName, result)
	return result, nil
}

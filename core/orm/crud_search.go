package orm

import (
	"context"
	"fmt"
	"time"
)

// SearchInterceptor allows addons to intercept and modify search domains.
type SearchInterceptor func(ctx context.Context, model string, domain [][]interface{}) ([][]interface{}, error)

var (
	SearchInterceptors []SearchInterceptor
)

// RegisterSearchInterceptor adds an interceptor to the global ORM search pipeline.
func RegisterSearchInterceptor(fn SearchInterceptor) {
	SearchInterceptors = append(SearchInterceptors, fn)
}

type searchPaging struct {
	orderBySQL string
	limit      int
	offset     int
}

func execSearchQuery(ctx context.Context, modelName string, domain [][]interface{}, paging *searchPaging) ([]map[string]interface{}, error) {
	ctx = ContextWithReadReplica(ctx, true)
	uid, whereClause, args, _, err := prepareSearchRead(ctx, modelName, domain)
	if err != nil {
		return nil, err
	}

	table, err := QuotedTableForModel(modelName)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	if paging != nil {
		placeholderIndex := len(args) + 1
		query += fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", paging.orderBySQL, placeholderIndex, placeholderIndex+1)
		args = append(args, paging.limit, paging.offset)
	}
	rows, err := QueryDB(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		recordMap, err := scanRowToMap(cols, rows)
		if err != nil {
			return nil, err
		}
		enrichRecordForRead(ctx, uid, modelName, recordMap)
		results = append(results, recordMap)
	}
	return results, rows.Err()
}

func prepareSearchRead(ctx context.Context, modelName string, domain [][]interface{}) (uid int, whereClause string, args []interface{}, domainOut [][]interface{}, err error) {
	if _, ok := Registry[modelName]; !ok {
		return 0, "", nil, nil, fmt.Errorf("model %s not found", modelName)
	}
	uid = SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return 0, "", nil, nil, err
	}
	for _, interceptor := range SearchInterceptors {
		domain, err = interceptor(ctx, modelName, domain)
		if err != nil {
			return 0, "", nil, nil, err
		}
	}
	whereClause, args, err = BuildWhereWithRecordRules(ctx, uid, modelName, "read", domain)
	if err != nil {
		return 0, "", nil, nil, err
	}
	return uid, whereClause, args, domain, nil
}

// Search finds records matching the criteria
func Search(ctx context.Context, modelName string, domain [][]interface{}) (results []map[string]interface{}, err error) {
	start := time.Now()
	defer func() {
		n := 0
		if results != nil {
			n = len(results)
		}
		logORMOperationKV(ctx, start, "search", modelName, err, "rows", n)
	}()
	results, err = execSearchQuery(ctx, modelName, domain, nil)
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

// SearchLimit returns up to limit rows for modelName matching domain, ordered by id.
// limit must be positive; otherwise it defaults to 500.
func SearchLimit(ctx context.Context, modelName string, domain [][]interface{}, limit int) (results []map[string]interface{}, err error) {
	return SearchPage(ctx, modelName, domain, limit, 0, "")
}

const (
	maxSearchLimit  = 500
	maxSearchOffset = 1_000_000
)

// SearchPage returns rows with database-level LIMIT/OFFSET.
// orderBy defaults to "id ASC". limit defaults/caps at 500; offset is clamped.
func SearchPage(ctx context.Context, modelName string, domain [][]interface{}, limit, offset int, orderBy string) (results []map[string]interface{}, err error) {
	start := time.Now()
	defer func() {
		n := 0
		if results != nil {
			n = len(results)
		}
		logORMOperationKV(ctx, start, "search_page", modelName, err, "rows", n, "limit", limit, "offset", offset)
	}()
	if limit <= 0 || limit > maxSearchLimit {
		limit = maxSearchLimit
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxSearchOffset {
		offset = maxSearchOffset
	}
	obSQL, err := ParseOrderByForModel(modelName, orderBy)
	if err != nil {
		return nil, err
	}
	return execSearchQuery(ctx, modelName, domain, &searchPaging{
		orderBySQL: obSQL,
		limit:      limit,
		offset:     offset,
	})
}

// SearchCount returns the number of rows matching domain (record rules applied).
func SearchCount(ctx context.Context, modelName string, domain [][]interface{}) (n int, err error) {
	start := time.Now()
	defer func() {
		logORMOperationKV(ctx, start, "search_count", modelName, err, "count", n)
	}()
	ctx = ContextWithReadReplica(ctx, true)
	_, whereClause, args, _, err := prepareSearchRead(ctx, modelName, domain)
	if err != nil {
		return 0, err
	}
	table, err := QuotedTableForModel(modelName)
	if err != nil {
		return 0, err
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereClause)
	err = QueryDB(ctx).QueryRowContext(ctx, query, args...).Scan(&n)
	return n, err
}

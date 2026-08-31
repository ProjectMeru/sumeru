package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func uiViewLookupLogErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

// FindUIDefaultView returns the highest-priority sys.view for a model and type.
// Record rules are applied in SQL like Search. Missing views return sql.ErrNoRows.
func FindUIDefaultView(ctx context.Context, modelName, viewType string) (result map[string]interface{}, err error) {
	start := time.Now()
	defer func() {
		logORMOperationKV(ctx, start, "find_ui_view", "sys.view", uiViewLookupLogErr(err), "target_model", modelName, "view_type", viewType, "found", result != nil)
	}()
	if _, ok := Registry["sys.view"]; !ok {
		return nil, fmt.Errorf("model sys.view not registered")
	}
	uid := SecurityUID(ctx)
	if !SecurityBypass(ctx) {
		if err := CheckModelAccess(ctx, uid, "sys.view", "read"); err != nil {
			return nil, err
		}
	}
	vt := strings.TrimSpace(strings.ToLower(viewType))
	return findUIDefaultViewByType(ctx, uid, modelName, vt)
}

func findUIDefaultViewByType(ctx context.Context, uid int, modelName, vt string) (map[string]interface{}, error) {
	domain := [][]interface{}{
		{"model", "=", modelName},
		{"type", "=", vt},
	}
	whereClause, args, err := BuildWhereWithRecordRules(ctx, uid, "sys.view", "read", domain)
	if err != nil {
		return nil, err
	}
	priCol, err := QuotedColumnForModel("sys.view", "priority")
	if err != nil {
		return nil, err
	}
	idCol, err := QuotedColumnForModel("sys.view", "id")
	if err != nil {
		return nil, err
	}
	tbl, err := QuotedTableForModel("sys.view")
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf(
		`SELECT * FROM %s WHERE %s ORDER BY %s DESC NULLS LAST, %s DESC LIMIT 1`,
		tbl, whereClause, priCol, idCol,
	)
	rows, err := DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	cols, _ := rows.Columns()
	return scanRowToMap(cols, rows)
}

// FindUIViewByName returns a sys.view row by unique name (typically the XML view id).
// Missing views return sql.ErrNoRows.
func FindUIViewByName(ctx context.Context, modelName, viewType, viewName string) (result map[string]interface{}, err error) {
	start := time.Now()
	defer func() {
		logORMOperationKV(ctx, start, "find_ui_view_by_name", "sys.view", uiViewLookupLogErr(err), "target_model", modelName, "view_type", viewType, "view_name", viewName, "found", result != nil)
	}()
	viewName = strings.TrimSpace(viewName)
	if viewName == "" {
		return nil, sql.ErrNoRows
	}
	vt := strings.TrimSpace(strings.ToLower(viewType))
	return SearchOne(ctx, "sys.view", map[string]interface{}{
		"model": modelName,
		"type":  vt,
		"name":  viewName,
	})
}

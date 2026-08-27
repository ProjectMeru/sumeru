package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Create inserts a new record into the database using the shared values pipeline.
func Create(ctx context.Context, model Model, values map[string]interface{}) (id int, err error) {
	start := time.Now()
	defer func() {
		logORMOperation(ctx, start, "create", model.ModelName(), err, map[string]interface{}{"resource_id": id})
	}()
	res, err := executeCreateMutation(ctx, model, values)
	return res.ID, err
}

// Upsert inserts or updates a record based on a unique field (usually 'name' or 'id').
func Upsert(ctx context.Context, model Model, values map[string]interface{}, conflictCol string) (id int, err error) {
	start := time.Now()
	defer func() {
		logORMOperation(ctx, start, "upsert", model.ModelName(), err, map[string]interface{}{"resource_id": id})
	}()
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, model.ModelName(), "create"); err != nil {
		return 0, err
	}
	prepared, err := PrepareValues(model, values, WriteOpCreate, PrepareOptions{StrictUnknown: !SecurityBypass(ctx)})
	if err != nil {
		return 0, err
	}
	if !SecurityBypass(ctx) {
		if err := CheckFieldWriteAccess(ctx, uid, model.ModelName(), prepared); err != nil {
			return 0, err
		}
	}
	if conflictCol != "" {
		if _, ok := prepared[conflictCol]; !ok {
			if v, ok := values[conflictCol]; ok {
				prepared[conflictCol] = v
			}
		}
	}
	if len(prepared) == 0 {
		return 0, fmt.Errorf("upsert requires at least one column")
	}

	table, err := QuotedTableForModel(model.ModelName())
	if err != nil {
		return 0, err
	}
	conflictQuoted, err := QuotedConflictColumn(model.ModelName(), conflictCol)
	if err != nil {
		return 0, err
	}

	cols, placeholders, args, err := buildInsertColumns(model.ModelName(), prepared)
	if err != nil {
		return 0, err
	}
	var updates []string
	for col := range prepared {
		if col == conflictCol {
			continue
		}
		qcol, err := QuotedColumnForModel(model.ModelName(), col)
		if err != nil {
			return 0, err
		}
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", qcol, qcol))
	}
	if len(updates) == 0 {
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", conflictQuoted, conflictQuoted))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING id",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "),
		conflictQuoted, strings.Join(updates, ", "))

	err = runMutationTx(ctx, func(tx TxWrapper) error {
		scanErr := tx.QueryRowContext(ctx, query, args...).Scan(&id)
		if scanErr == sql.ErrNoRows {
			id = 0
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		if SecurityBypass(ctx) {
			AppendAuditTx(ctx, tx, "upsert", model.ModelName(), int64(id), nil, prepared, "source=module_sync")
		}
		return nil
	})
	return id, err
}

func buildInsertColumns(modelName string, prepared map[string]interface{}) (cols []string, placeholders []string, args []interface{}, err error) {
	i := 1
	for col, val := range prepared {
		qcol, err := QuotedColumnForModel(modelName, col)
		if err != nil {
			return nil, nil, nil, err
		}
		cols = append(cols, qcol)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		i++
	}
	return cols, placeholders, args, nil
}

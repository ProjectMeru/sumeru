package orm

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/errcode"
)

func insertPreparedOnTx(ctx context.Context, tx TxWrapper, model Model, prepared map[string]interface{}) (int, error) {
	cols, placeholders, args, err := buildInsertColumns(model.ModelName(), prepared)
	if err != nil {
		return 0, err
	}
	table, err := QuotedTableForModel(model.ModelName())
	if err != nil {
		return 0, err
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	var id int
	err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
	return id, err
}

// insertRawOnTx inserts side-effect rows without ACL checks (caller must pass bypass ctx).
func insertRawOnTx(ctx context.Context, tx TxWrapper, model Model, values map[string]interface{}) (int, error) {
	prepared, err := PrepareValues(model, values, WriteOpCreate, PrepareOptions{
		StrictUnknown:     false,
		AllowPasswordHash: passwordHashWriteAllowed(ctx),
	})
	if err != nil {
		return 0, err
	}
	if len(prepared) == 0 {
		return 0, fmt.Errorf("insert requires at least one column")
	}
	return insertPreparedOnTx(ctx, tx, model, prepared)
}

// execSideEffectOnTx runs fn on tx without aborting the parent transaction on failure.
func execSideEffectOnTx(ctx context.Context, tx TxWrapper, modelName, operation string, fn func() error) {
	if tx == nil {
		if err := fn(); err != nil {
			logSideEffectWarn(ctx, operation, modelName, err)
		}
		return
	}
	if _, err := tx.ExecContext(ctx, "SAVEPOINT sp_orm_side"); err != nil {
		logSideEffectWarn(ctx, operation, modelName, err)
		return
	}
	if err := fn(); err != nil {
		_, _ = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT sp_orm_side")
		logSideEffectWarn(ctx, operation, modelName, err)
		return
	}
	if _, err := tx.ExecContext(ctx, "RELEASE SAVEPOINT sp_orm_side"); err != nil {
		logSideEffectWarn(ctx, operation, modelName, err)
	}
}

func lockRowsForDomain(ctx context.Context, tx TxWrapper, table, whereSQL string, args []interface{}) ([]map[string]interface{}, error) {
	lockQ := fmt.Sprintf(`SELECT * FROM %s WHERE %s FOR UPDATE`, table, whereSQL)
	rows, err := tx.QueryContext(ctx, lockQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var locked []map[string]interface{}
	for rows.Next() {
		m, err := scanRowToMap(cols, rows)
		if err != nil {
			return nil, err
		}
		locked = append(locked, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(locked) == 0 {
		return nil, fmt.Errorf("access denied or record not found")
	}
	return locked, nil
}

func insertSideEffectRow(ctx context.Context, tx TxWrapper, registryKey string, vals map[string]interface{}) error {
	inst, ok := Registry[registryKey]
	if !ok || inst == nil {
		return fmt.Errorf("model %q not registered", registryKey)
	}
	bypass := ContextWithBypass(ctx, true)
	if tx != nil {
		var err error
		execSideEffectOnTx(bypass, tx, registryKey, "insert_side_effect", func() error {
			_, err = insertRawOnTx(bypass, tx, inst, vals)
			return err
		})
		return err
	}
	if DB == nil {
		return fmt.Errorf("no database")
	}
	_, err := Create(bypass, inst, vals)
	if err != nil {
		applog.WarnCode(ctx, errcode.InternalError, "Side effect insert failed", applog.Event{
			Component: "orm",
			Operation: "insert_side_effect",
			Status:    "partial",
			Context:   map[string]interface{}{"resource": registryKey},
			Err:       err,
		})
	}
	return err
}

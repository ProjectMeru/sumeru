package orm

import (
	"context"
	"fmt"
	"time"
)

// Unlink deletes a record by ID using domain-scoped UnlinkWhere.
func Unlink(ctx context.Context, modelName string, id int) (err error) {
	start := time.Now()
	defer func() {
		logORMOperation(ctx, start, "delete", modelName, err, map[string]interface{}{"resource_id": id})
	}()
	if 0 >= id {
		return fmt.Errorf("invalid id")
	}
	_, err = UnlinkWhere(ctx, modelName, [][]interface{}{{"id", "=", id}})
	return err
}

// UnlinkWhere deletes rows matching domain inside a mandatory transaction.
func UnlinkWhere(ctx context.Context, modelName string, domain [][]interface{}) (n int64, err error) {
	start := time.Now()
	defer func() {
		logORMOperation(ctx, start, "delete", modelName, err, map[string]interface{}{"rows": n})
	}()
	res, err := executeDeleteMutation(ctx, modelName, domain)
	return res.RowsAffected, err
}

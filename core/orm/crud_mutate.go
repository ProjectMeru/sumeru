package orm

import (
	"context"
	"fmt"
	"strings"
)

// runMutationTx runs fn inside a database transaction and commits on success.
func runMutationTx(ctx context.Context, fn func(tx TxWrapper) error) error {
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

type mutationKind string

const (
	mutationCreate mutationKind = "create"
	mutationUpdate mutationKind = "update"
	mutationDelete mutationKind = "delete"
)

type mutationResult struct {
	ID              int
	RowsAffected    int64
	PendingEventIDs []int
	EventName       string
}

func executeCreateMutation(ctx context.Context, model Model, values map[string]interface{}) (mutationResult, error) {
	var result mutationResult
	prepared, uid, err := prepareCreateWrite(ctx, model, values, PrepareOptions{StrictUnknown: true})
	if err != nil {
		return result, err
	}
	result.EventName = EventRecordCreated
	err = runMutationTx(ctx, func(tx TxWrapper) error {
		id, err := insertPreparedOnTx(ctx, tx, model, prepared)
		if err != nil {
			return err
		}
		result.ID = id
		if shouldEmitSideEffects(ctx, model.ModelName()) {
			emitSideEffectsOnTx(ctx, tx, model.ModelName(), uid, []sideEffectRow{{
				Action:    "create",
				EventName: EventRecordCreated,
				ResID:     int64(id),
				After:     prepared,
			}})
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if shouldEmitSideEffects(ctx, model.ModelName()) {
		publishRecordEvents(ctx, result.EventName, uid, model.ModelName(), []int{result.ID})
	}
	return result, nil
}

func mutationAccessCheck(ctx context.Context, modelName, op string) (uid int, model Model, err error) {
	uid = SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, op); err != nil {
		return 0, nil, err
	}
	model, err = requireRegisteredModel(modelName)
	if err != nil {
		return 0, nil, err
	}
	return uid, model, nil
}

func executeUpdateMutation(ctx context.Context, modelName string, domain [][]interface{}, values map[string]interface{}) (mutationResult, error) {
	var result mutationResult
	uid, inst, err := mutationAccessCheck(ctx, modelName, "write")
	if err != nil {
		return result, err
	}
	prepared, err := PrepareValues(inst, values, WriteOpWrite, PrepareOptions{
		StrictUnknown:     false,
		AllowPasswordHash: passwordHashWriteAllowed(ctx),
	})
	if err != nil {
		return result, err
	}
	if err := RejectVirtualWrites(inst, values); err != nil {
		return result, err
	}
	if err := CheckFieldWriteAccess(ctx, uid, modelName, prepared); err != nil {
		return result, err
	}
	if len(prepared) == 0 {
		return result, nil
	}
	table, err := QuotedTableForModel(modelName)
	if err != nil {
		return result, err
	}
	securedSQL, args, err := BuildWhereWithRecordRules(ctx, uid, modelName, "write", domain)
	if err != nil {
		return result, err
	}
	result.EventName = EventRecordUpdated
	var sideRows []sideEffectRow
	err = runMutationTx(ctx, func(tx TxWrapper) error {
		locked, err := lockRowsForDomain(ctx, tx, table, securedSQL, args)
		if err != nil {
			return err
		}
		if len(locked) == 0 {
			return fmt.Errorf("access denied or record not found")
		}
		if len(locked) == 1 {
			merged := mergeRecordMap(locked[0], prepared)
			if err := MergeStoredComputes(ctx, modelName, merged); err != nil {
				return err
			}
			for k, v := range merged {
				if fd := FieldDef(modelName, k); fd != nil && fd.ComputeStore {
					prepared[k] = v
				}
			}
		}
		for _, before := range locked {
			if newState, ok := prepared["state"].(string); ok {
				oldState := AsString(before["state"])
				if newState != oldState {
					rid, _ := CoerceInt64(before["id"])
					if err := CanWorkflowTransition(ctx, WorkflowTransitionInput{
						Model: modelName, RecordID: int(rid), FromState: oldState, ToState: newState, UID: uid,
					}); err != nil {
						return err
					}
				}
			}
			merged := mergeRecordMap(before, prepared)
			if err := runConstraints(ctx, modelName, merged); err != nil {
				return err
			}
			if err := CheckRecordRules(ctx, uid, modelName, "write", merged); err != nil {
				return err
			}
		}
		var setClauses []string
		var setArgs []interface{}
		placeholderIndex := 1
		for k, v := range prepared {
			qcol, err := QuotedColumnForModel(modelName, k)
			if err != nil {
				return err
			}
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", qcol, placeholderIndex))
			setArgs = append(setArgs, v)
			placeholderIndex++
		}
		shiftedWhere, err := shiftPlaceholders(securedSQL, len(setArgs)+1)
		if err != nil {
			return err
		}
		allArgs := append(setArgs, args...)
		updQ := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, table, strings.Join(setClauses, ", "), shiftedWhere)
		res, err := tx.ExecContext(ctx, updQ, allArgs...)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return fmt.Errorf("access denied or record not found")
		}
		result.RowsAffected = n
		if shouldEmitSideEffects(ctx, modelName) {
			for _, before := range locked {
				rid, _ := CoerceInt64(before["id"])
				merged := mergeRecordMap(before, prepared)
				sideRows = append(sideRows, sideEffectRow{
					Action:    "write",
					EventName: EventRecordUpdated,
					ResID:     rid,
					Before:    before,
					After:     merged,
				})
				result.PendingEventIDs = append(result.PendingEventIDs, int(rid))
			}
			emitSideEffectsOnTx(ctx, tx, modelName, uid, sideRows)
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if len(result.PendingEventIDs) > 0 {
		publishRecordEvents(ctx, result.EventName, uid, modelName, result.PendingEventIDs)
	}
	return result, nil
}

func executeDeleteMutation(ctx context.Context, modelName string, domain [][]interface{}) (mutationResult, error) {
	var result mutationResult
	uid, _, err := mutationAccessCheck(ctx, modelName, "unlink")
	if err != nil {
		return result, err
	}
	table, err := QuotedTableForModel(modelName)
	if err != nil {
		return result, err
	}
	securedSQL, args, err := BuildWhereWithRecordRules(ctx, uid, modelName, "unlink", domain)
	if err != nil {
		return result, err
	}
	result.EventName = EventRecordDeleted
	var sideRows []sideEffectRow
	err = runMutationTx(ctx, func(tx TxWrapper) error {
		locked, err := lockRowsForDomain(ctx, tx, table, securedSQL, args)
		if err != nil {
			return err
		}
		for _, recordMap := range locked {
			if err := CheckRecordRules(ctx, uid, modelName, "unlink", recordMap); err != nil {
				return err
			}
		}
		delQ := fmt.Sprintf(`DELETE FROM %s WHERE %s`, table, securedSQL)
		res, err := tx.ExecContext(ctx, delQ, args...)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return fmt.Errorf("access denied or record not found")
		}
		result.RowsAffected = n
		if shouldEmitSideEffects(ctx, modelName) {
			for _, recordMap := range locked {
				recordID, _ := CoerceInt64(recordMap["id"])
				sideRows = append(sideRows, sideEffectRow{
					Action:    "unlink",
					EventName: EventRecordDeleted,
					ResID:     recordID,
					Before:    recordMap,
				})
				result.PendingEventIDs = append(result.PendingEventIDs, int(recordID))
			}
			emitSideEffectsOnTx(ctx, tx, modelName, uid, sideRows)
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if len(result.PendingEventIDs) > 0 {
		publishRecordEvents(ctx, result.EventName, uid, modelName, result.PendingEventIDs)
	}
	return result, nil
}

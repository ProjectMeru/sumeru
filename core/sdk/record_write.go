package sdk

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

// BinaryAttachmentInput is the file stored beside a record.
type BinaryAttachmentInput = orm.CreateBinaryAttachmentInput

// EventRecordDeleted is the name published after a row is removed.
const EventRecordDeleted = orm.EventRecordDeleted

// Create inserts one row.
func Create(ctx context.Context, modelName string, values map[string]interface{}) (int, error) {
	inst, err := registeredModel(modelName)
	if err != nil {
		return 0, err
	}
	return orm.Create(ctx, inst, values)
}

// Upsert inserts or updates on a unique conflict column.
func Upsert(ctx context.Context, modelName string, values map[string]interface{}, conflictCol string) (int, error) {
	inst, err := registeredModel(modelName)
	if err != nil {
		return 0, err
	}
	return orm.Upsert(ctx, inst, values, conflictCol)
}

// Write sets values on the given ids. One id and many ids share one domain update when more than one id is passed.
func Write(ctx context.Context, modelName string, ids []int, values map[string]interface{}) (int64, error) {
	if 0 == len(ids) {
		return 0, fmt.Errorf("%s: record ids required", modelName)
	}
	if len(ids) == 1 {
		if err := orm.UpdateRecordByID(ctx, modelName, ids[0], values); err != nil {
			return 0, err
		}
		return 1, nil
	}
	return orm.Update(ctx, modelName, idDomain(ids), values)
}

// SearchWrite updates every row matching domain and returns the row count.
func SearchWrite(ctx context.Context, modelName string, domain [][]interface{}, values map[string]interface{}) (int64, error) {
	return orm.Update(ctx, modelName, domain, values)
}

// Unlink deletes the given ids.
func Unlink(ctx context.Context, modelName string, ids []int) (int64, error) {
	if 0 == len(ids) {
		return 0, fmt.Errorf("%s: record ids required", modelName)
	}
	if len(ids) == 1 {
		if err := orm.Unlink(ctx, modelName, ids[0]); err != nil {
			return 0, err
		}
		return 1, nil
	}
	return orm.UnlinkWhere(ctx, modelName, idDomain(ids))
}

// Copy duplicates one row.
func Copy(ctx context.Context, modelName string, id int) (int, error) {
	return orm.Copy(ctx, modelName, id)
}

// NameCreate inserts a row from a many2one completion label.
func NameCreate(ctx context.Context, modelName, name string) (int, error) {
	return orm.NameCreate(ctx, modelName, name)
}

// CreateBinaryAttachment stores a file and creates the attachment row.
func CreateBinaryAttachment(ctx context.Context, in BinaryAttachmentInput) (int, error) {
	return orm.CreateBinaryAttachment(ctx, in)
}

// RegisterBeforeDelete sets the single pre-delete hook.
func RegisterBeforeDelete(fn func(ctx context.Context, model string, id int, before map[string]interface{})) {
	orm.RegisterBeforeDelete(fn)
}

func registeredModel(modelName string) (orm.Model, error) {
	inst := orm.RegistryModel(modelName)
	if inst == nil {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	return inst, nil
}

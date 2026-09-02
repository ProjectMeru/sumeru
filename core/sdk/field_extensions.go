package sdk

import "sumeru/core/orm"

// RegisterFieldMerger registers a callback that merges extra fields into model field maps.
func RegisterFieldMerger(fn func(modelName string, fieldDefs map[string]orm.FieldDefinition)) {
	orm.RegisterFieldMerger(fn)
}

package orm

import "sync"

var (
	fieldMergerMu sync.RWMutex
	fieldMergers  []func(modelName string, fieldDefs map[string]FieldDefinition)
)

// RegisterFieldMerger adds a callback that merges extra field definitions into model field maps.
// Enterprise Studio registers manual sys.field rows here; core stays agnostic.
func RegisterFieldMerger(fn func(modelName string, fieldDefs map[string]FieldDefinition)) {
	if fn == nil {
		return
	}
	fieldMergerMu.Lock()
	defer fieldMergerMu.Unlock()
	fieldMergers = append(fieldMergers, fn)
}

// ClearFieldMergers removes registered mergers (tests).
func ClearFieldMergers() {
	fieldMergerMu.Lock()
	defer fieldMergerMu.Unlock()
	fieldMergers = nil
}

func applyFieldMergers(modelName string, fieldDefs map[string]FieldDefinition) {
	fieldMergerMu.RLock()
	list := append([]func(modelName string, fieldDefs map[string]FieldDefinition){}, fieldMergers...)
	fieldMergerMu.RUnlock()
	for _, fn := range list {
		fn(modelName, fieldDefs)
	}
}

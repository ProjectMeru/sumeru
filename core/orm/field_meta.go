package orm

// IsVirtualField reports whether a field has no physical SQL column.
func IsVirtualField(f FieldDefinition) bool {
	if f.Virtual {
		return true
	}
	if f.Related != "" && !f.RelatedStore {
		return true
	}
	if f.Compute != "" && !f.ComputeStore {
		return true
	}
	return false
}

func fieldDefinitionsByName(model Model) map[string]FieldDefinition {
	fieldDefs := map[string]FieldDefinition{}
	if model == nil {
		return fieldDefs
	}
	for _, field := range model.Fields() {
		if field.Name == "" || field.Name == "id" {
			continue
		}
		fieldDefs[field.Name] = field
	}
	applyFieldMergers(model.ModelName(), fieldDefs)
	return fieldDefs
}

// FieldDef returns the field definition for model.field, or nil.
func FieldDef(modelName, fieldName string) *FieldDefinition {
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return nil
	}
	for i := range inst.Fields() {
		f := inst.Fields()[i]
		if f.Name == fieldName {
			cp := f
			return &cp
		}
	}
	return nil
}

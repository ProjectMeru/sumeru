package form

// WritableValues exposes writableValues to external tests.
func (f *Form) WritableValues() map[string]interface{} { return f.writableValues() }

// PutDraft injects a raw draft value bypassing Set's guards, for filter tests.
func (f *Form) PutDraft(field string, v interface{}) { f.values[field] = v }

package orm

// ModelRegistered reports whether modelName is in the ORM registry.
func ModelRegistered(modelName string) bool {
	_, ok := Registry[modelName]
	return ok
}

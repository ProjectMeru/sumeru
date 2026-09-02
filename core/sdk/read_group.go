package sdk

import "sumeru/core/orm"

// RegisterReadGroupWrapper sets the optional ReadGroup wrapper (enterprise ops).
func RegisterReadGroupWrapper(w orm.ReadGroupWrapper) {
	orm.RegisterReadGroupWrapper(w)
}

// ClearReadGroupWrapper removes the wrapper (tests).
func ClearReadGroupWrapper() {
	orm.ClearReadGroupWrapper()
}

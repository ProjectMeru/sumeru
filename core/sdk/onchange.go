package sdk

import "sumeru/core/orm"

// OnchangeWarning is an optional message shown after a field change.
type OnchangeWarning = orm.OnchangeWarning

// OnchangeResult is the value patch returned to the form.
type OnchangeResult = orm.OnchangeResult

// RegisterOnchange wires a field onchange handler.
func RegisterOnchange(model, field string, fn orm.OnchangeHandler) {
	orm.RegisterOnchange(model, field, fn)
}

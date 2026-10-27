package sdk

import (
	"context"

	"sumeru/core/orm"
)

// OnchangeWarning is an optional message shown after a field change.
type OnchangeWarning struct {
	Title   string
	Message string
}

// OnchangeResult is the value patch returned to the form.
type OnchangeResult struct {
	Value   map[string]interface{}
	Warning *OnchangeWarning
	Domain  map[string]interface{}
}

// RegisterOnchange wires a field onchange handler.
func RegisterOnchange(model, field string, fn func(ctx context.Context, values map[string]interface{}, field string) (OnchangeResult, error)) {
	orm.RegisterOnchange(model, field, func(ctx context.Context, values map[string]interface{}, field string) (orm.OnchangeResult, error) {
		res, err := fn(ctx, values, field)
		if err != nil {
			return orm.OnchangeResult{}, err
		}
		out := orm.OnchangeResult{Value: res.Value, Domain: res.Domain}
		if res.Warning != nil {
			out.Warning = &orm.OnchangeWarning{Title: res.Warning.Title, Message: res.Warning.Message}
		}
		return out, nil
	})
}

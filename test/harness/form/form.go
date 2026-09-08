// Package form provides a server-side pseudo-form test helper inspired by
// Odoo's odoo.tests.Form. It simulates a browser form by applying field
// defaults, onchange handlers and computed fields to an in-memory draft, and
// validates/writes the draft through the ORM on Save.
//
// Scope (v1): scalar fields (char/text/integer/float/boolean/selection/many2one),
// required validation, readonly enforcement, onchange, and computed fields.
// Relational x2m commands (one2many/many2many) are not supported yet.
package form

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// Form is an in-memory draft of a model record that mirrors a form view.
type Form struct {
	ctx    context.Context
	model  orm.Model
	values map[string]interface{}
	id     int
	saved  bool
}

// New creates a Form for a fresh record of model, pre-populating field defaults
// and refreshing computed fields.
func New(ctx context.Context, model orm.Model) (*Form, error) {
	if model == nil || model.ModelName() == "" {
		return nil, fmt.Errorf("form: nil or unnamed model")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	f := &Form{ctx: ctx, model: model, values: map[string]interface{}{}}
	for _, fd := range model.Fields() {
		if fd.Name == "" || fd.Name == "id" {
			continue
		}
		if v, ok := defaultValue(fd); ok {
			f.values[fd.Name] = v
		}
	}
	_ = orm.ApplyComputes(ctx, model.ModelName(), f.values)
	return f, nil
}

// NewModel creates a Form for a fresh record, looking the model up in orm.Registry.
func NewModel(ctx context.Context, modelName string) (*Form, error) {
	m := orm.RegistryModel(modelName)
	if m == nil {
		return nil, fmt.Errorf("form: model %q not registered", modelName)
	}
	return New(ctx, m)
}

// Set assigns a draft field value, then runs the field's onchange (if any) and
// refreshes computed fields. Readonly, computed and related fields are rejected;
// relational x2m fields are not supported yet.
func (f *Form) Set(field string, v interface{}) error {
	fd := fieldDef(f.model, field)
	if fd == nil {
		return fmt.Errorf("unknown field %q on model %s", field, f.model.ModelName())
	}
	if fd.Type == orm.One2Many || fd.Type == orm.Many2Many {
		return fmt.Errorf("field %q on %s is a relational (x2m) field: not supported yet", field, f.model.ModelName())
	}
	if fd.Readonly {
		return fmt.Errorf("field %q on %s is read-only", field, f.model.ModelName())
	}
	if fd.Compute != "" {
		return fmt.Errorf("field %q on %s is computed", field, f.model.ModelName())
	}
	if fd.Related != "" {
		return fmt.Errorf("field %q on %s is related", field, f.model.ModelName())
	}
	if orm.IsVirtualField(*fd) {
		return fmt.Errorf("field %q on %s is read-only", field, f.model.ModelName())
	}
	f.values[field] = v
	f.applyOnchange(field)
	return nil
}

// Get returns the draft value for field, refreshing computed fields first.
func (f *Form) Get(field string) interface{} {
	_ = orm.ApplyComputes(f.ctx, f.model.ModelName(), f.values)
	if v, ok := f.values[field]; ok {
		return v
	}
	if fd := fieldDef(f.model, field); fd != nil {
		if v, ok := defaultValue(*fd); ok {
			return v
		}
	}
	return nil
}

// Validate runs the ORM's create-time field validation (type coercion and
// required checks) on the draft without writing to the database.
func (f *Form) Validate() error {
	_, err := orm.PrepareValues(f.model, f.writableValues(), orm.WriteOpCreate, orm.PrepareOptions{StrictUnknown: false})
	return err
}

// Save writes the draft to the database: create on first call, update after.
func (f *Form) Save() error {
	writable := f.writableValues()
	if !f.saved {
		id, err := orm.Create(f.ctx, f.model, writable)
		if err != nil {
			return err
		}
		f.id = id
		f.saved = true
		return nil
	}
	return orm.UpdateRecordByID(f.ctx, f.model.ModelName(), f.id, writable)
}

// Record reads the saved record back from the database.
func (f *Form) Record() (map[string]interface{}, error) {
	if !f.saved {
		return nil, fmt.Errorf("form: record not saved yet")
	}
	return orm.SearchOne(f.ctx, f.model.ModelName(), map[string]interface{}{"id": f.id})
}

// ID returns the saved record id, or 0 when not saved yet.
func (f *Form) ID() int { return f.id }

// Saved reports whether Save has completed.
func (f *Form) Saved() bool { return f.saved }

// Draft returns a shallow copy of the current draft values.
func (f *Form) Draft() map[string]interface{} {
	out := make(map[string]interface{}, len(f.values))
	for k, v := range f.values {
		out[k] = v
	}
	return out
}

func (f *Form) applyOnchange(field string) {
	if orm.HasOnchange(f.model.ModelName(), field) {
		if res, err := orm.RunOnchange(f.ctx, f.model.ModelName(), field, f.values); err == nil {
			for k, v := range res.Value {
				f.values[k] = v
			}
		}
	}
	_ = orm.ApplyComputes(f.ctx, f.model.ModelName(), f.values)
}

// writableValues returns only the fields the ORM accepts in create/write
// payloads: it drops virtual, computed, related and relational x2m fields.
func (f *Form) writableValues() map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range f.values {
		if k == "id" {
			continue
		}
		fd := fieldDef(f.model, k)
		if fd == nil {
			continue
		}
		if fd.Type == orm.One2Many || fd.Type == orm.Many2Many {
			continue
		}
		if fd.Compute != "" || fd.Related != "" || orm.IsVirtualField(*fd) {
			continue
		}
		out[k] = v
	}
	return out
}

func fieldDef(model orm.Model, name string) *orm.FieldDefinition {
	if model == nil {
		return nil
	}
	for _, fd := range model.Fields() {
		if fd.Name == name {
			cp := fd
			return &cp
		}
	}
	return nil
}

// defaultValue returns the field's static default, skipping runtime tokens
// (current_user/current_company/uuid) that the ORM resolves at write time.
func defaultValue(fd orm.FieldDefinition) (interface{}, bool) {
	if fd.DefaultVal == nil {
		return nil, false
	}
	if s, ok := fd.DefaultVal.(string); ok {
		switch strings.TrimSpace(s) {
		case "current_user", "current_company", "uuid":
			return nil, false
		}
	}
	return fd.DefaultVal, true
}

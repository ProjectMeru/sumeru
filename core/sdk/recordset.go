package sdk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"sumeru/core/orm"
)

// ActionOpen describes a form opened after a record method.
type ActionOpen struct {
	Model    string
	ActionID int
	ViewType string
	RecordID int
	Target   string
}

// ActionResult is what a record method returns to the client.
type ActionResult struct {
	Redirect string
	Open     *ActionOpen
	Close    bool
}

// MethodFunc handles a button for the records it was called on.
type MethodFunc func(ctx context.Context, rs Recordset) (ActionResult, error)

// Recordset is one record or many of the same model.
type Recordset struct {
	ctx   context.Context
	model string
	ids   []int
	vals  map[string]string
}

// Browse builds a set. An empty id list is an empty set.
func Browse(ctx context.Context, model string, ids ...int) Recordset {
	copied := append([]int(nil), ids...)
	return Recordset{ctx: ctx, model: model, ids: copied}
}

// Model returns the model name.
func (r Recordset) Model() string { return r.model }

// IDs returns a copy of the id list.
func (r Recordset) IDs() []int { return append([]int(nil), r.ids...) }

// Len returns how many records are in the set.
func (r Recordset) Len() int { return len(r.ids) }

// Vals returns the button form values captured with the call.
func (r Recordset) Vals() map[string]string { return r.vals }

// EnsureOne returns the only id. Zero or several ids is an error.
func (r Recordset) EnsureOne() (int, error) {
	if len(r.ids) != 1 {
		return 0, fmt.Errorf("%s: expected one record, got %d", r.model, len(r.ids))
	}
	return r.ids[0], nil
}

// Read loads rows for this set. The predicate helpers below use the same read.
func (r Recordset) Read(fields []string) ([]map[string]interface{}, error) {
	return Read(r.ctx, r.model, r.ids, fields)
}

// Write updates every record in the set.
func (r Recordset) Write(values map[string]interface{}) (int64, error) {
	return Write(r.ctx, r.model, r.ids, values)
}

// Unlink deletes every record in the set.
func (r Recordset) Unlink() (int64, error) {
	return Unlink(r.ctx, r.model, r.ids)
}

// Filtered keeps rows for which fn returns true. fn is applied in memory after one read.
func (r Recordset) Filtered(fn func(row map[string]interface{}) bool) (Recordset, error) {
	if fn == nil {
		return r, nil
	}
	rows, err := r.Read(nil)
	if err != nil {
		return Recordset{}, err
	}
	var ids []int
	for _, row := range rows {
		if !fn(row) {
			continue
		}
		id, ok := CoerceInt64(row["id"])
		if ok && id > 0 {
			ids = append(ids, int(id))
		}
	}
	return Browse(r.ctx, r.model, ids...), nil
}

// Mapped returns one cell from each row already loaded by a read. It does not add a SQL filter.
func (r Recordset) Mapped(field string) ([]interface{}, error) {
	rows, err := r.Read([]string{field})
	if err != nil {
		return nil, err
	}
	out := make([]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, row[field])
	}
	return out, nil
}

// RegisterMethod stores a record method on the kernel object-action registry.
// One id and many ids call the same function.
func RegisterMethod(model, name string, fn MethodFunc) {
	if fn == nil {
		return
	}
	orm.RegisterObjectActionIDs(model, name, func(ctx context.Context, model string, ids []int, vals map[string]string) (string, error) {
		rs := Browse(ctx, model, ids...)
		rs.vals = vals
		result, err := fn(ctx, rs)
		if err != nil {
			return "", err
		}
		return result.legacyRedirect(), nil
	})
}

// RegisterObjectAction stores a one-id button handler on the kernel registry.
func RegisterObjectAction(model, method string, fn func(ctx context.Context, model string, id int, vals map[string]string) (string, error)) {
	orm.RegisterObjectAction(model, method, fn)
}

func (r ActionResult) legacyRedirect() string {
	if r.Close {
		return orm.ActionCloseToken()
	}
	if r.Open != nil && r.Open.Model != "" && r.Open.RecordID > 0 {
		viewType := r.Open.ViewType
		if viewType == "" {
			viewType = "form"
		}
		target := r.Open.Target
		if target == "" {
			target = "dialog"
		}
		q := url.Values{}
		q.Set("model", r.Open.Model)
		q.Set("id", strconv.Itoa(r.Open.RecordID))
		q.Set("view_type", viewType)
		q.Set("target", target)
		if r.Open.ActionID > 0 {
			q.Set("action", strconv.Itoa(r.Open.ActionID))
		}
		return "/web?" + q.Encode()
	}
	return r.Redirect
}

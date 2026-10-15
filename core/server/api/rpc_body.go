package api

import (
	"encoding/json"
	"fmt"
)

type builtRPCBody struct {
	args   json.RawMessage
	kwargs json.RawMessage
}

// BuildRPCRequestBody maps a flat JSON body to legacy args/kwargs for dispatchRPC.
func BuildRPCRequestBody(method string, raw []byte) (builtRPCBody, error) {
	if len(raw) == 0 {
		return builtRPCBody{}, newRPCError(CodeInvalidJSON, "empty body", nil)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return builtRPCBody{}, newRPCError(CodeInvalidJSON, fmt.Sprintf("invalid json: %v", err), nil)
	}
	kwargs := extractContextKwargs(fields)
	switch method {
	case "search":
		return buildSearch(fields, kwargs)
	case "search_read":
		return buildSearchRead(fields, kwargs)
	case "read":
		return buildRead(fields, kwargs)
	case "create":
		return buildCreate(fields, kwargs)
	case "write":
		return buildWrite(fields, kwargs)
	case "unlink":
		return buildUnlink(fields, kwargs)
	case "create_many":
		return buildCreateMany(fields, kwargs)
	case "write_many":
		return buildWriteMany(fields, kwargs)
	case "unlink_many":
		return buildUnlinkMany(fields, kwargs)
	case "read_group":
		return buildReadGroup(fields, kwargs)
	case "onchange":
		return buildOnchange(fields, kwargs)
	case "call":
		return buildCall(fields, kwargs)
	default:
		return builtRPCBody{}, newRPCError(CodeMethodNotAllowed, fmt.Sprintf("unsupported method %q", method), map[string]interface{}{"method": method})
	}
}

func extractContextKwargs(fields map[string]json.RawMessage) json.RawMessage {
	raw, ok := fields["context"]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	kw := map[string]json.RawMessage{"context": raw}
	for _, key := range []string{"limit", "offset", "order"} {
		if v, ok := fields[key]; ok {
			kw[key] = v
		}
	}
	out, err := json.Marshal(kw)
	if err != nil {
		return nil
	}
	return out
}

func mergeKwargs(base json.RawMessage, extra map[string]json.RawMessage) (json.RawMessage, error) {
	m := map[string]interface{}{}
	if len(base) > 0 && string(base) != "null" {
		if err := json.Unmarshal(base, &m); err != nil {
			return nil, err
		}
	}
	for k, v := range extra {
		var decoded interface{}
		if err := json.Unmarshal(v, &decoded); err != nil {
			return nil, err
		}
		m[k] = decoded
	}
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

func requireField(fields map[string]json.RawMessage, name string) (json.RawMessage, error) {
	raw, ok := fields[name]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return nil, newRPCError(CodeValidationError, fmt.Sprintf("%s is required", name), map[string]interface{}{"field": name})
	}
	return raw, nil
}

func buildSearch(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	domain, err := requireField(fields, "domain")
	if err != nil {
		return builtRPCBody{}, err
	}
	args, err := json.Marshal([]json.RawMessage{domain})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildSearchRead(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	domain, err := requireField(fields, "domain")
	if err != nil {
		return builtRPCBody{}, err
	}
	fieldsRaw, err := requireField(fields, "fields")
	if err != nil {
		return builtRPCBody{}, err
	}
	extra := map[string]json.RawMessage{}
	for _, key := range []string{"limit", "offset", "order"} {
		if v, ok := fields[key]; ok {
			extra[key] = v
		}
	}
	kw, err := mergeKwargs(kwargs, extra)
	if err != nil {
		return builtRPCBody{}, newRPCError(CodeInvalidArgs, err.Error(), nil)
	}
	args, err := json.Marshal([]json.RawMessage{domain, fieldsRaw})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kw}, nil
}

func buildRead(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	ids, err := requireField(fields, "ids")
	if err != nil {
		return builtRPCBody{}, err
	}
	parts := []json.RawMessage{ids}
	if f, ok := fields["fields"]; ok && len(f) > 0 && string(f) != "null" {
		parts = append(parts, f)
	}
	args, err := json.Marshal(parts)
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildCreate(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	values, err := requireField(fields, "values")
	if err != nil {
		return builtRPCBody{}, err
	}
	args, err := json.Marshal([]json.RawMessage{values})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildWrite(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	ids, err := requireField(fields, "ids")
	if err != nil {
		return builtRPCBody{}, err
	}
	values, err := requireField(fields, "values")
	if err != nil {
		return builtRPCBody{}, err
	}
	args, err := json.Marshal([]json.RawMessage{ids, values})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildUnlink(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	ids, err := requireField(fields, "ids")
	if err != nil {
		return builtRPCBody{}, err
	}
	args, err := json.Marshal([]json.RawMessage{ids})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildCreateMany(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	values, err := requireField(fields, "values")
	if err != nil {
		// allow list at top level as values array
		if v, ok := fields["records"]; ok {
			values = v
		} else {
			return builtRPCBody{}, err
		}
	}
	args, err := json.Marshal([]json.RawMessage{values})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildWriteMany(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	return buildWrite(fields, kwargs)
}

func buildUnlinkMany(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	return buildUnlink(fields, kwargs)
}

func buildReadGroup(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	spec, ok := fields["spec"]
	if !ok || len(spec) == 0 {
		specBytes, err := json.Marshal(map[string]json.RawMessage{
			"domain":  fields["domain"],
			"groupby": fields["groupby"],
			"fields":  fields["fields"],
		})
		if err != nil {
			return builtRPCBody{}, err
		}
		spec = specBytes
	}
	args, err := json.Marshal([]json.RawMessage{spec})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildOnchange(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	values, err := requireField(fields, "values")
	if err != nil {
		return builtRPCBody{}, err
	}
	field, err := requireField(fields, "field")
	if err != nil {
		return builtRPCBody{}, err
	}
	args, err := json.Marshal([]json.RawMessage{values, field})
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

func buildCall(fields map[string]json.RawMessage, kwargs json.RawMessage) (builtRPCBody, error) {
	idRaw, err := requireField(fields, "id")
	if err != nil {
		return builtRPCBody{}, err
	}
	method, err := requireField(fields, "method")
	if err != nil {
		return builtRPCBody{}, err
	}
	parts := []json.RawMessage{idRaw, method}
	if v, ok := fields["vals"]; ok && len(v) > 0 && string(v) != "null" {
		parts = append(parts, v)
	}
	args, err := json.Marshal(parts)
	if err != nil {
		return builtRPCBody{}, err
	}
	return builtRPCBody{args: args, kwargs: kwargs}, nil
}

package api

import (
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func parseDomainArg(raw json.RawMessage) ([][]interface{}, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	return orm.ParseDomainJSON(string(raw))
}

func normArgs(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("[]")
	}
	return raw
}

func parseLimitOffset(kwargs json.RawMessage) (limit int, offset int) {
	limit, offset = 500, 0
	if len(kwargs) == 0 || string(kwargs) == "null" {
		return limit, offset
	}
	var m map[string]interface{}
	if err := json.Unmarshal(kwargs, &m); err != nil {
		return limit, offset
	}
	if v, ok := m["limit"]; ok {
		if f, ok := toFloat(v); ok {
			limit = int(f)
		}
	}
	if v, ok := m["offset"]; ok {
		if f, ok := toFloat(v); ok {
			offset = int(f)
		}
	}
	if limit <= 0 {
		limit = 500
	}
	return limit, offset
}

func toFloat(v interface{}) (float64, bool) {
	return orm.CoerceFloat64(v)
}

func parseArgsArray(args json.RawMessage) ([]json.RawMessage, error) {
	args = normArgs(args)
	var arr []json.RawMessage
	if err := json.Unmarshal(args, &arr); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args must be a JSON array: %v", err), nil)
	}
	return arr, nil
}

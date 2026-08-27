package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

func rpcOnchange(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	var payload []json.RawMessage
	if err := json.Unmarshal(args, &payload); err != nil || len(payload) < 2 {
		return nil, newRPCError(CodeInvalidArgs, "onchange requires args[0] values object and args[1] field name", map[string]interface{}{"method": "onchange"})
	}
	var values map[string]interface{}
	if err := json.Unmarshal(payload[0], &values); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[0] values: %v", err), map[string]interface{}{"method": "onchange"})
	}
	var field string
	if err := json.Unmarshal(payload[1], &field); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[1] field: %v", err), map[string]interface{}{"method": "onchange"})
	}
	field = strings.TrimSpace(field)
	if field == "" {
		return nil, newRPCError(CodeInvalidArgs, "field name is required", map[string]interface{}{"method": "onchange"})
	}
	result, err := orm.RunOnchange(ctx, model, field, values)
	if err != nil {
		return nil, newRPCError(CodeMethodNotAllowed, err.Error(), map[string]interface{}{"method": "onchange", "field": field})
	}
	return result, nil
}

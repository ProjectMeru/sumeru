package api

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func rpcWrite(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, newRPCError(CodeInvalidArgs, "write requires args[0] ids and args[1] values", map[string]interface{}{"method": "write"})
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[0] ids: %v", err), map[string]interface{}{"method": "write"})
	}
	if err := capRPCIDs(ids); err != nil {
		return nil, err
	}
	var vals map[string]interface{}
	if err := json.Unmarshal(arr[1], &vals); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[1] values: %v", err), map[string]interface{}{"method": "write"})
	}
	for _, id := range ids {
		if err := orm.UpdateRecordByID(ctx, model, id, vals); err != nil {
			return nil, err
		}
	}
	return true, nil
}

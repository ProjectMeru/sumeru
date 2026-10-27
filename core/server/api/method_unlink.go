package api

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func rpcUnlink(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, newRPCError(CodeInvalidArgs, "unlink requires args[0] ids", map[string]interface{}{"method": "unlink"})
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[0] ids: %v", err), map[string]interface{}{"method": "unlink"})
	}
	if err := capRPCIDs(ids); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if err := orm.Unlink(ctx, model, id); err != nil {
			return nil, err
		}
	}
	return true, nil
}

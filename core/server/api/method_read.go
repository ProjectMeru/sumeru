package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func rpcRead(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, newRPCError(CodeInvalidArgs, "read requires args[0] ids", map[string]interface{}{"method": "read"})
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("ids: %v", err), map[string]interface{}{"method": "read"})
	}
	if err := capRPCIDs(ids); err != nil {
		return nil, err
	}
	var fields []string
	if len(arr) >= 2 && len(arr[1]) > 0 && string(arr[1]) != "null" {
		if err := json.Unmarshal(arr[1], &fields); err != nil {
			return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[1] fields: %v", err), map[string]interface{}{"method": "read"})
		}
	}
	if len(fields) > 0 {
		if err := orm.CheckFieldReadAccess(ctx, orm.SecurityUID(ctx), model, fields); err != nil {
			return nil, err
		}
	}
	var out []map[string]interface{}
	var missing []int
	for _, id := range ids {
		rec, err := orm.SearchOne(ctx, model, map[string]interface{}{"id": id})
		if err != nil {
			if err == sql.ErrNoRows {
				missing = append(missing, id)
				continue
			}
			return nil, err
		}
		if len(fields) > 0 {
			out = append(out, projectFields([]map[string]interface{}{rec}, fields)[0])
		} else {
			out = append(out, rec)
		}
	}
	if len(missing) > 0 {
		return nil, newRPCError(CodeNotFound, "record(s) not found", map[string]interface{}{"missing_ids": missing})
	}
	return out, nil
}

package api

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/engine/swcmeta"
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
	if 0 == len(ids) {
		return []map[string]interface{}{}, nil
	}
	idValues := make([]interface{}, len(ids))
	for i, id := range ids {
		idValues[i] = id
	}
	rows, err := orm.Search(ctx, model, [][]interface{}{{"id", "in", idValues}})
	if err != nil {
		return nil, err
	}
	byID := make(map[int]map[string]interface{}, len(rows))
	for _, row := range rows {
		id, ok := orm.CoerceInt64(row["id"])
		if !ok {
			continue
		}
		byID[int(id)] = row
	}
	var missing []int
	out := make([]map[string]interface{}, 0, len(ids))
	uid := orm.SecurityUID(ctx)
	for _, id := range ids {
		rec, ok := byID[id]
		if !ok {
			missing = append(missing, id)
			continue
		}
		row := rec
		if len(fields) > 0 {
			row = projectFields([]map[string]interface{}{rec}, fields)[0]
		}
		orm.RedactRecordForRead(ctx, uid, model, row)
		out = append(out, row)
	}
	if len(missing) > 0 {
		return nil, newRPCError(CodeNotFound, "record(s) not found", map[string]interface{}{"missing_ids": missing})
	}
	swcmeta.EnrichMany2OneNames(ctx, model, out)
	return out, nil
}

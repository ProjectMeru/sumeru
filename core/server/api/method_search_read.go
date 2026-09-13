package api

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func rpcSearchRead(ctx context.Context, model string, args, kwargs json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, newRPCError(CodeInvalidArgs, "search_read requires args[0] domain and args[1] fields", map[string]interface{}{"method": "search_read"})
	}
	domain, err := parseDomainArg(arr[0])
	if err != nil {
		return nil, newRPCError(CodeInvalidArgs, err.Error(), map[string]interface{}{"method": "search_read", "hint": "args[0] domain"})
	}
	domain = orm.SubstituteDomainUID(domain, orm.UIDFromContext(ctx))
	ctx = orm.ContextWithReadReplica(ctx, true)
	var fields []string
	if err := json.Unmarshal(arr[1], &fields); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[1] fields: %v", err), map[string]interface{}{"method": "search_read"})
	}
	limit, offset := parseLimitOffset(kwargs)
	uid := orm.UIDFromContext(ctx)
	rows, err := orm.SearchPage(ctx, model, domain, limit, offset, "")
	if err != nil {
		return nil, err
	}
	orm.RedactSearchResults(ctx, uid, model, rows)
	if len(fields) > 0 {
		if err := orm.CheckFieldReadAccess(ctx, uid, model, fields); err != nil {
			return nil, err
		}
	}
	return projectFields(rows, fields), nil
}

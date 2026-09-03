package api

import (
	"context"
	"encoding/json"
)

func StatusForCodeForTest(code string) int { return statusForCode(code) }

func NewRPCErrorForTest(code, msg string, details map[string]interface{}) error {
	return newRPCError(code, msg, details)
}

type CodedErrorForTest = codedError

func RPCErrorCodeForTest(err error) (string, bool) {
	ce, ok := err.(*codedError)
	if !ok {
		return "", false
	}
	return ce.code, true
}

func NormArgsForTest(raw json.RawMessage) json.RawMessage { return normArgs(raw) }

func ParseArgsArrayForTest(args json.RawMessage) ([]json.RawMessage, error) {
	return parseArgsArray(args)
}

func ParseDomainArgForTest(raw json.RawMessage) ([][]interface{}, error) {
	return parseDomainArg(raw)
}

func ParseLimitOffsetForTest(kwargs json.RawMessage) (limit int, offset int) {
	return parseLimitOffset(kwargs)
}

func ToFloatForTest(v interface{}) (float64, bool) { return toFloat(v) }

func ProjectFieldsForTest(rows []map[string]interface{}, fields []string) []map[string]interface{} {
	return projectFields(rows, fields)
}

func ValidateKwargsForTest(raw json.RawMessage) error { return validateKwargs(raw) }

func RPCReadForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcRead(ctx, model, args)
}

func RPCCreateForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcCreate(ctx, model, args)
}

func RPCWriteForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcWrite(ctx, model, args)
}

func RPCUnlinkForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcUnlink(ctx, model, args)
}

func RPCCreateManyForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcCreateMany(ctx, model, args)
}

func RPCWriteManyForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcWriteMany(ctx, model, args)
}

func RPCUnlinkManyForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcUnlinkMany(ctx, model, args)
}

func RPCOnchangeForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcOnchange(ctx, model, args)
}

func RPCReadGroupForTest(ctx context.Context, model string, args json.RawMessage, kwargs json.RawMessage) (interface{}, error) {
	return rpcReadGroup(ctx, model, args, kwargs)
}

func RPCCallForTest(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	return rpcCall(ctx, model, args)
}

func RPCSearchForTest(ctx context.Context, model string, args, kwargs json.RawMessage) (interface{}, error) {
	return rpcSearch(ctx, model, args, kwargs)
}

func RPCSearchReadForTest(ctx context.Context, model string, args, kwargs json.RawMessage) (interface{}, error) {
	return rpcSearchRead(ctx, model, args, kwargs)
}

func CapRPCIDsForTest(ids []int) error { return capRPCIDs(ids) }

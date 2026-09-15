// Package api exposes JSON-RPC over HTTP for authenticated sessions or API keys.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// PublicMethods is the allowlist of model methods callable via RPC.
var PublicMethods = map[string]bool{
	"search":       true,
	"search_read":  true,
	"read":         true,
	"create":       true,
	"write":        true,
	"unlink":       true,
	"create_many":  true,
	"write_many":   true,
	"unlink_many":  true,
	"read_group":   true,
	"onchange":     true,
	"call":         true,
}

// rpcRequest is the JSON body for POST /api/rpc.
// If model is empty and params is set, params is unmarshalled again (JSON-RPC wrapper).
type rpcRequest struct {
	Model  string          `json:"model"`
	Method string          `json:"method"`
	Args   json.RawMessage `json:"args"`
	Kwargs json.RawMessage `json:"kwargs"`
	Params json.RawMessage `json:"params"`
}

// dispatchRPC executes a model method call with the same security context as the web UI (uid in ctx).
func dispatchRPC(ctx context.Context, body []byte) (interface{}, error) {
	if len(body) == 0 {
		return nil, newRPCError(CodeInvalidJSON, "empty body", nil)
	}
	var in rpcRequest
	if err := json.Unmarshal(body, &in); err != nil {
		return nil, newRPCError(CodeInvalidJSON, fmt.Sprintf("invalid json: %v", err), nil)
	}
	if strings.TrimSpace(in.Model) == "" && len(in.Params) > 0 {
		if err := json.Unmarshal(in.Params, &in); err != nil {
			return nil, newRPCError(CodeInvalidJSON, fmt.Sprintf("invalid params: %v", err), nil)
		}
	}
	if err := validateKwargs(in.Kwargs); err != nil {
		return nil, err
	}
	model := strings.TrimSpace(in.Model)
	method := strings.TrimSpace(in.Method)
	if model == "" {
		return nil, newRPCError(CodeValidationError, "model is required", map[string]interface{}{"field": "model"})
	}
	if method == "" {
		return nil, newRPCError(CodeValidationError, "method is required", map[string]interface{}{"field": "method"})
	}
	if !PublicMethods[method] {
		return nil, newRPCError(CodeMethodNotAllowed, fmt.Sprintf("method %q is not a public RPC method", method), map[string]interface{}{"method": method})
	}
	if _, ok := orm.Registry[model]; !ok {
		return nil, newRPCError(CodeModelNotFound, fmt.Sprintf("unknown model %q", model), map[string]interface{}{"model": model})
	}
	if orm.UIDFromContext(ctx) <= 0 {
		return nil, newRPCError(CodeUnauthorized, "authentication required", nil)
	}
	if err := orm.EnforcePortalRPCModel(ctx, model); err != nil {
		return nil, newRPCError(CodeAccessDenied, err.Error(), map[string]interface{}{"model": model})
	}

	switch method {
	case "search":
		return rpcSearch(ctx, model, in.Args, in.Kwargs)
	case "search_read":
		return rpcSearchRead(ctx, model, in.Args, in.Kwargs)
	case "read":
		return rpcRead(ctx, model, in.Args)
	case "create":
		return rpcCreate(ctx, model, in.Args)
	case "write":
		return rpcWrite(ctx, model, in.Args)
	case "unlink":
		return rpcUnlink(ctx, model, in.Args)
	case "create_many":
		return rpcCreateMany(ctx, model, in.Args)
	case "write_many":
		return rpcWriteMany(ctx, model, in.Args)
	case "unlink_many":
		return rpcUnlinkMany(ctx, model, in.Args)
	case "read_group":
		return rpcReadGroup(ctx, model, in.Args, in.Kwargs)
	case "onchange":
		return rpcOnchange(ctx, model, in.Args)
	case "call":
		return rpcCall(ctx, model, in.Args)
	default:
		return nil, newRPCError(CodeMethodNotAllowed, fmt.Sprintf("unsupported method %q", method), map[string]interface{}{"method": method})
	}
}

func validateKwargs(raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var probe interface{}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return newRPCError(CodeInvalidArgs, fmt.Sprintf("kwargs: %v", err), nil)
	}
	if _, ok := probe.(map[string]interface{}); !ok {
		return newRPCError(CodeInvalidArgs, "kwargs must be a JSON object", nil)
	}
	return nil
}

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

func rpcReadGroup(ctx context.Context, model string, args json.RawMessage, kwargs json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, newRPCError(CodeInvalidArgs, "read_group requires args[0] spec object", map[string]interface{}{"method": "read_group"})
	}
	specBytes, _ := json.Marshal(arr[0])
	var spec struct {
		Domain  json.RawMessage `json:"domain"`
		GroupBy []string        `json:"groupby"`
		Fields  []struct {
			Name    string `json:"name"`
			Field   string `json:"field"`
			Measure string `json:"measure"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("spec: %v", err), map[string]interface{}{"method": "read_group"})
	}
	var domain [][]interface{}
	if len(spec.Domain) > 0 && string(spec.Domain) != "null" {
		domain, err = orm.ParseDomainJSON(string(spec.Domain))
		if err != nil {
			return nil, newRPCError(CodeInvalidArgs, err.Error(), map[string]interface{}{"method": "read_group"})
		}
	}
	rg := orm.ReadGroupSpec{GroupBy: spec.GroupBy}
	for _, f := range spec.Fields {
		rg.Fields = append(rg.Fields, orm.ReadGroupField{Name: f.Name, Field: f.Field, Measure: f.Measure})
	}
	if 0 == len(rg.Fields) && len(rg.GroupBy) > 0 {
		rg.Fields = []orm.ReadGroupField{{Name: "__count", Field: "id", Measure: "count"}}
	}
	return orm.ReadGroup(ctx, model, domain, rg)
}

func rpcCall(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, newRPCError(CodeInvalidArgs, "call requires args[0] id and args[1] method", map[string]interface{}{"method": "call"})
	}
	var idF float64
	if err := json.Unmarshal(arr[0], &idF); err != nil || int(idF) <= 0 {
		return nil, newRPCError(CodeInvalidArgs, "args[0] must be record id", map[string]interface{}{"method": "call"})
	}
	var method string
	if err := json.Unmarshal(arr[1], &method); err != nil || strings.TrimSpace(method) == "" {
		return nil, newRPCError(CodeInvalidArgs, "args[1] must be method name", map[string]interface{}{"method": "call"})
	}
	vals := map[string]string{}
	if len(arr) >= 3 {
		var m map[string]interface{}
		if err := json.Unmarshal(arr[2], &m); err == nil {
			for k, v := range m {
				vals[k] = fmt.Sprint(v)
			}
		}
	}
	redirect, err := orm.RunObjectAction(ctx, model, int(idF), method, vals)
	if err != nil {
		return nil, err
	}
	return orm.EncodeActionResult(redirect), nil
}

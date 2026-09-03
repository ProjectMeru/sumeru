package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	_ "sumeru/core/ormmodels"
	"sumeru/core/orm"
	"sumeru/core/server/api"
)



func testCtx() context.Context {
	return orm.ContextWithUID(context.Background(), 1)
}

func TestDispatchRPC_validation(t *testing.T) {
	cases := []struct {
		ctx        context.Context
		body       string
		wantStatus int
		wantCode   string
	}{
		{testCtx(), `{"model":"sys.session","method":"search","args":{}}`, http.StatusBadRequest, api.CodeInvalidArgs},
		{context.Background(), `{"model":"sys.session","method":"search","args":[]}`, http.StatusUnauthorized, api.CodeUnauthorized},
		{testCtx(), `{"model":"sys.session","method":"search_read","args":[[]]}`, http.StatusBadRequest, api.CodeInvalidArgs},
		{testCtx(), `{"model":"sys.session","method":"read","args":[]}`, http.StatusBadRequest, api.CodeInvalidArgs},
	}
	for _, tc := range cases {
		resp, status := api.Dispatch(tc.ctx, []byte(tc.body))
		if status != tc.wantStatus {
			t.Fatalf("body=%s status=%d want=%d", tc.body, status, tc.wantStatus)
		}
		if resp.Error == nil || resp.Error.Code != tc.wantCode {
			t.Fatalf("body=%s error=%+v want=%s", tc.body, resp.Error, tc.wantCode)
		}
	}
}

func TestRPCMethods_argValidation(t *testing.T) {
	ctx := testCtx()
	model := "sys.session"

	t.Run("rpcRead", func(t *testing.T) {
		_, err := api.RPCReadForTest(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCReadForTest(ctx, model, json.RawMessage(`[[1],{}]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		ids := make([]int, 501)
		for i := range ids {
			ids[i] = 1
		}
		raw, _ := json.Marshal([]interface{}{ids})
		_, err = api.RPCReadForTest(ctx, model, raw)
		assertCoded(t, err, api.CodeInvalidArgs)
		out, err := api.RPCReadForTest(ctx, model, json.RawMessage(`[[]]`))
		if err != nil {
			t.Fatalf("empty ids: %v", err)
		}
		if rows, ok := out.([]map[string]interface{}); !ok || len(rows) != 0 {
			t.Fatalf("out = %v", out)
		}
		out, err = api.RPCReadForTest(ctx, model, json.RawMessage(`[[],["sid"]]`))
		if err != nil {
			t.Fatalf("empty ids with fields: %v", err)
		}
		if rows, ok := out.([]map[string]interface{}); !ok || len(rows) != 0 {
			t.Fatalf("out = %v", out)
		}
	})

	t.Run("rpcCreate", func(t *testing.T) {
		_, err := api.RPCCreateForTest(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCreateForTest(ctx, model, json.RawMessage(`[[]]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCreateForTest(ctx, model, json.RawMessage(`[{"sid":"test"}]`))
		if err == nil {
			t.Fatal("expected error without database")
		}
		_, err = api.RPCCreateForTest(ctx, "no.such.model", json.RawMessage(`[{"x":1}]`))
		assertCoded(t, err, api.CodeModelNotFound)
	})

	t.Run("rpcWrite", func(t *testing.T) {
		_, err := api.RPCWriteForTest(ctx, model, json.RawMessage(`[[]]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCWriteForTest(ctx, model, json.RawMessage(`[{},{}]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCWriteForTest(ctx, model, json.RawMessage(`[[1],[]]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		v, err := api.RPCWriteForTest(ctx, model, json.RawMessage(`[[],{"sid":"x"}]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
	})

	t.Run("rpcUnlink", func(t *testing.T) {
		_, err := api.RPCUnlinkForTest(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCUnlinkForTest(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		v, err := api.RPCUnlinkForTest(ctx, model, json.RawMessage(`[[]]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
	})

	t.Run("rpcCreateMany", func(t *testing.T) {
		_, err := api.RPCCreateManyForTest(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCreateManyForTest(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCreateManyForTest(ctx, model, json.RawMessage(`[[{"sid":"a"}]]`))
		if err == nil {
			t.Fatal("expected error without database")
		}
		rows := make([]string, 501)
		for i := range rows {
			rows[i] = `{}`
		}
		_, err = api.RPCCreateManyForTest(ctx, model, json.RawMessage(`[`+strings.Join(rows, ",")+`]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCreateManyForTest(ctx, "no.such.model", json.RawMessage(`[[{"sid":"a"}]]`))
		assertCoded(t, err, api.CodeModelNotFound)
	})

	t.Run("rpcWriteMany", func(t *testing.T) {
		_, err := api.RPCWriteManyForTest(ctx, model, json.RawMessage(`[[1]]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		v, err := api.RPCWriteManyForTest(ctx, model, json.RawMessage(`[[],{}]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
		_, err = api.RPCWriteManyForTest(ctx, model, json.RawMessage(`[[1],[]]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		ids := make([]int, 501)
		raw, _ := json.Marshal([]interface{}{ids, map[string]string{"sid": "x"}})
		_, err = api.RPCWriteManyForTest(ctx, model, raw)
		assertCoded(t, err, api.CodeInvalidArgs)
	})

	t.Run("rpcUnlinkMany", func(t *testing.T) {
		_, err := api.RPCUnlinkManyForTest(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		v, err := api.RPCUnlinkManyForTest(ctx, model, json.RawMessage(`[[]]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
		_, err = api.RPCUnlinkManyForTest(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		ids := make([]int, 501)
		raw, _ := json.Marshal([]interface{}{ids})
		_, err = api.RPCUnlinkManyForTest(ctx, model, raw)
		assertCoded(t, err, api.CodeInvalidArgs)
	})

	t.Run("rpcOnchange", func(t *testing.T) {
		_, err := api.RPCOnchangeForTest(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCOnchangeForTest(ctx, model, json.RawMessage(`[[],1]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCOnchangeForTest(ctx, model, json.RawMessage(`[{},""]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCOnchangeForTest(ctx, model, json.RawMessage(`[{"sid":"x"},"sid"]`))
		if err == nil {
			t.Fatal("expected error from RunOnchange without handler")
		}
		_, err = api.RPCOnchangeForTest(ctx, model, json.RawMessage(`{}`))
		assertCoded(t, err, api.CodeInvalidArgs)
	})

	t.Run("rpcReadGroup", func(t *testing.T) {
		_, err := api.RPCReadGroupForTest(ctx, model, json.RawMessage(`[]`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCReadGroupForTest(ctx, model, json.RawMessage(`[{"domain":"bad","groupby":["id"]}]`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCReadGroupForTest(ctx, model, json.RawMessage(`[1]`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
	})

	t.Run("rpcCall", func(t *testing.T) {
		_, err := api.RPCCallForTest(ctx, model, json.RawMessage(`[1]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCallForTest(ctx, model, json.RawMessage(`["x","m"]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCallForTest(ctx, model, json.RawMessage(`[1,""]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCCallForTest(ctx, model, json.RawMessage(`[1,1]`))
		assertCoded(t, err, api.CodeInvalidArgs)
		func() {
			defer func() { _ = recover() }()
			_, _ = api.RPCCallForTest(ctx, model, json.RawMessage(`[1,"noop",{"k":"v"}]`))
		}()
	})

	t.Run("rpcSearch", func(t *testing.T) {
		_, err := api.RPCSearchForTest(ctx, model, json.RawMessage(`{}`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCSearchForTest(ctx, model, json.RawMessage(`[["bad"]]`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
	})

	t.Run("rpcSearchRead", func(t *testing.T) {
		_, err := api.RPCSearchReadForTest(ctx, model, json.RawMessage(`[[]]`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCSearchReadForTest(ctx, model, json.RawMessage(`[[],{}]`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
		_, err = api.RPCSearchReadForTest(ctx, model, json.RawMessage(`[[["bad"]],["id"]]`), nil)
		assertCoded(t, err, api.CodeInvalidArgs)
	})

	t.Run("capRPCIDs", func(t *testing.T) {
		ids := make([]int, 501)
		err := api.CapRPCIDsForTest(ids)
		assertCoded(t, err, api.CodeInvalidArgs)
	})
}

func assertCoded(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	got, ok := api.RPCErrorCodeForTest(err)
	if !ok || got != code {
		t.Fatalf("err = %v; want code %q", err, code)
	}
}

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	_ "sumeru/core/ormmodels"
	"sumeru/core/orm"
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
		{testCtx(), `{"model":"sys.session","method":"search","args":{}}`, http.StatusBadRequest, CodeInvalidArgs},
		{context.Background(), `{"model":"sys.session","method":"search","args":[]}`, http.StatusUnauthorized, CodeUnauthorized},
		{testCtx(), `{"model":"sys.session","method":"search_read","args":[[]]}`, http.StatusBadRequest, CodeInvalidArgs},
		{testCtx(), `{"model":"sys.session","method":"read","args":[]}`, http.StatusBadRequest, CodeInvalidArgs},
	}
	for _, tc := range cases {
		resp, status := Dispatch(tc.ctx, []byte(tc.body))
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
		_, err := rpcRead(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcRead(ctx, model, json.RawMessage(`[[1],{}]`))
		assertCoded(t, err, CodeInvalidArgs)
		ids := make([]int, 501)
		for i := range ids {
			ids[i] = 1
		}
		raw, _ := json.Marshal([]interface{}{ids})
		_, err = rpcRead(ctx, model, raw)
		assertCoded(t, err, CodeInvalidArgs)
		out, err := rpcRead(ctx, model, json.RawMessage(`[[]]`))
		if err != nil {
			t.Fatalf("empty ids: %v", err)
		}
		if rows, ok := out.([]map[string]interface{}); !ok || len(rows) != 0 {
			t.Fatalf("out = %v", out)
		}
		out, err = rpcRead(ctx, model, json.RawMessage(`[[],["sid"]]`))
		if err != nil {
			t.Fatalf("empty ids with fields: %v", err)
		}
		if rows, ok := out.([]map[string]interface{}); !ok || len(rows) != 0 {
			t.Fatalf("out = %v", out)
		}
	})

	t.Run("rpcCreate", func(t *testing.T) {
		_, err := rpcCreate(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCreate(ctx, model, json.RawMessage(`[[]]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCreate(ctx, model, json.RawMessage(`[{"sid":"test"}]`))
		if err == nil {
			t.Fatal("expected error without database")
		}
		_, err = rpcCreate(ctx, "no.such.model", json.RawMessage(`[{"x":1}]`))
		assertCoded(t, err, CodeModelNotFound)
	})

	t.Run("rpcWrite", func(t *testing.T) {
		_, err := rpcWrite(ctx, model, json.RawMessage(`[[]]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcWrite(ctx, model, json.RawMessage(`[{},{}]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcWrite(ctx, model, json.RawMessage(`[[1],[]]`))
		assertCoded(t, err, CodeInvalidArgs)
		v, err := rpcWrite(ctx, model, json.RawMessage(`[[],{"sid":"x"}]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
	})

	t.Run("rpcUnlink", func(t *testing.T) {
		_, err := rpcUnlink(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcUnlink(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, CodeInvalidArgs)
		v, err := rpcUnlink(ctx, model, json.RawMessage(`[[]]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
	})

	t.Run("rpcCreateMany", func(t *testing.T) {
		_, err := rpcCreateMany(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCreateMany(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCreateMany(ctx, model, json.RawMessage(`[[{"sid":"a"}]]`))
		if err == nil {
			t.Fatal("expected error without database")
		}
		rows := make([]string, 501)
		for i := range rows {
			rows[i] = `{}`
		}
		_, err = rpcCreateMany(ctx, model, json.RawMessage(`[`+strings.Join(rows, ",")+`]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCreateMany(ctx, "no.such.model", json.RawMessage(`[[{"sid":"a"}]]`))
		assertCoded(t, err, CodeModelNotFound)
	})

	t.Run("rpcWriteMany", func(t *testing.T) {
		_, err := rpcWriteMany(ctx, model, json.RawMessage(`[[1]]`))
		assertCoded(t, err, CodeInvalidArgs)
		v, err := rpcWriteMany(ctx, model, json.RawMessage(`[[],{}]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
		_, err = rpcWriteMany(ctx, model, json.RawMessage(`[[1],[]]`))
		assertCoded(t, err, CodeInvalidArgs)
		ids := make([]int, 501)
		raw, _ := json.Marshal([]interface{}{ids, map[string]string{"sid": "x"}})
		_, err = rpcWriteMany(ctx, model, raw)
		assertCoded(t, err, CodeInvalidArgs)
	})

	t.Run("rpcUnlinkMany", func(t *testing.T) {
		_, err := rpcUnlinkMany(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, CodeInvalidArgs)
		v, err := rpcUnlinkMany(ctx, model, json.RawMessage(`[[]]`))
		if err != nil || v != true {
			t.Fatalf("empty ids: v=%v err=%v", v, err)
		}
		_, err = rpcUnlinkMany(ctx, model, json.RawMessage(`[{}]`))
		assertCoded(t, err, CodeInvalidArgs)
		ids := make([]int, 501)
		raw, _ := json.Marshal([]interface{}{ids})
		_, err = rpcUnlinkMany(ctx, model, raw)
		assertCoded(t, err, CodeInvalidArgs)
	})

	t.Run("rpcOnchange", func(t *testing.T) {
		_, err := rpcOnchange(ctx, model, json.RawMessage(`[]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcOnchange(ctx, model, json.RawMessage(`[[],1]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcOnchange(ctx, model, json.RawMessage(`[{},""]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcOnchange(ctx, model, json.RawMessage(`[{"sid":"x"},"sid"]`))
		if err == nil {
			t.Fatal("expected error from RunOnchange without handler")
		}
		_, err = rpcOnchange(ctx, model, json.RawMessage(`{}`))
		assertCoded(t, err, CodeInvalidArgs)
	})

	t.Run("rpcReadGroup", func(t *testing.T) {
		_, err := rpcReadGroup(ctx, model, json.RawMessage(`[]`), nil)
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcReadGroup(ctx, model, json.RawMessage(`[{"domain":"bad","groupby":["id"]}]`), nil)
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcReadGroup(ctx, model, json.RawMessage(`[1]`), nil)
		assertCoded(t, err, CodeInvalidArgs)
	})

	t.Run("rpcCall", func(t *testing.T) {
		_, err := rpcCall(ctx, model, json.RawMessage(`[1]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCall(ctx, model, json.RawMessage(`["x","m"]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCall(ctx, model, json.RawMessage(`[1,""]`))
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcCall(ctx, model, json.RawMessage(`[1,1]`))
		assertCoded(t, err, CodeInvalidArgs)
		func() {
			defer func() { _ = recover() }()
			_, _ = rpcCall(ctx, model, json.RawMessage(`[1,"noop",{"k":"v"}]`))
		}()
	})

	t.Run("rpcSearch", func(t *testing.T) {
		_, err := rpcSearch(ctx, model, json.RawMessage(`{}`), nil)
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcSearch(ctx, model, json.RawMessage(`[["bad"]]`), nil)
		assertCoded(t, err, CodeInvalidArgs)
	})

	t.Run("rpcSearchRead", func(t *testing.T) {
		_, err := rpcSearchRead(ctx, model, json.RawMessage(`[[]]`), nil)
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcSearchRead(ctx, model, json.RawMessage(`[[],{}]`), nil)
		assertCoded(t, err, CodeInvalidArgs)
		_, err = rpcSearchRead(ctx, model, json.RawMessage(`[[["bad"]],["id"]]`), nil)
		assertCoded(t, err, CodeInvalidArgs)
	})

	t.Run("capRPCIDs", func(t *testing.T) {
		ids := make([]int, 501)
		err := capRPCIDs(ids)
		assertCoded(t, err, CodeInvalidArgs)
	})
}

func assertCoded(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	ce, ok := err.(*codedError)
	if !ok || ce.code != code {
		t.Fatalf("err = %v; want code %q", err, code)
	}
}

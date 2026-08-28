package api_test

import (
	"context"
	"net/http"
	"testing"

	_ "sumeru/core/ormmodels"
	"sumeru/core/orm"
	"sumeru/core/server/api"
)

func authCtx() context.Context {
	return orm.ContextWithUID(context.Background(), 1)
}

func assertDispatch(t *testing.T, ctx context.Context, body string, wantStatus int, wantCode string) {
	t.Helper()
	resp, status := api.Dispatch(ctx, []byte(body))
	if status != wantStatus {
		t.Fatalf("status = %d; want %d (body=%s)", status, wantStatus, body)
	}
	if wantStatus == http.StatusOK {
		if !resp.OK {
			t.Fatalf("ok = false; want true (body=%s)", body)
		}
		return
	}
	if resp.OK {
		t.Fatalf("ok = true; want false (body=%s)", body)
	}
	if resp.Error == nil {
		t.Fatal("error is nil")
	}
	if resp.Error.Code != wantCode {
		t.Fatalf("error.code = %q; want %q (msg=%q)", resp.Error.Code, wantCode, resp.Error.Message)
	}
}

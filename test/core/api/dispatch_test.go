package api_test

import (
	"sumeru/core/server/api"
	"context"
	"net/http"
	"testing"
	"sumeru/core/orm"
)



func TestDispatchRPC_edgeCases(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 1)

	t.Run("kwargs invalid json", func(t *testing.T) {
		body := `{"model":"sys.session","method":"search","args":[],"kwargs":"not-json"}`
		resp, status := api.Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != api.CodeInvalidArgs {
			t.Fatalf("status=%d code=%q resp=%+v", status, resp.Error.Code, resp)
		}
	})

	t.Run("trim model and method", func(t *testing.T) {
		body := `{"model":"  sys.session  ","method":" read ","args":[]}`
		resp, status := api.Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != api.CodeInvalidArgs {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("params wrapper validation", func(t *testing.T) {
		body := `{"params":{"model":"sys.session","method":"read","args":[]}}`
		resp, status := api.Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != api.CodeInvalidArgs {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("whitespace only model", func(t *testing.T) {
		body := `{"model":"   ","method":"search","args":[]}`
		resp, status := api.Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != api.CodeValidationError {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("whitespace only method", func(t *testing.T) {
		body := `{"model":"sys.session","method":"   ","args":[]}`
		resp, status := api.Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != api.CodeValidationError {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("unsupported public method default", func(t *testing.T) {
		api.PublicMethods["__test_only__"] = true
		defer delete(api.PublicMethods, "__test_only__")
		resp, status := api.Dispatch(ctx, []byte(`{"model":"sys.session","method":"__test_only__","args":[]}`))
		if status != http.StatusForbidden || resp.Error.Code != api.CodeMethodNotAllowed {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})
}

func TestValidateKwargs(t *testing.T) {
	if err := api.ValidateKwargsForTest(nil); err != nil {
		t.Fatal(err)
	}
	if err := api.ValidateKwargsForTest([]byte("null")); err != nil {
		t.Fatal(err)
	}
	err := api.ValidateKwargsForTest([]byte(`[]`))
	if err == nil {
		t.Fatal("expected error for array kwargs")
	}
}

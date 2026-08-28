package api

import (
	"context"
	"net/http"
	"testing"

	_ "sumeru/core/ormmodels"
	"sumeru/core/orm"
)

func TestDispatchRPC_edgeCases(t *testing.T) {
	ctx := orm.ContextWithUID(context.Background(), 1)

	t.Run("kwargs invalid json", func(t *testing.T) {
		body := `{"model":"sys.session","method":"search","args":[],"kwargs":"not-json"}`
		resp, status := Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != CodeInvalidArgs {
			t.Fatalf("status=%d code=%q resp=%+v", status, resp.Error.Code, resp)
		}
	})

	t.Run("trim model and method", func(t *testing.T) {
		body := `{"model":"  sys.session  ","method":" read ","args":[]}`
		resp, status := Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != CodeInvalidArgs {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("params wrapper validation", func(t *testing.T) {
		body := `{"params":{"model":"sys.session","method":"read","args":[]}}`
		resp, status := Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != CodeInvalidArgs {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("whitespace only model", func(t *testing.T) {
		body := `{"model":"   ","method":"search","args":[]}`
		resp, status := Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != CodeValidationError {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("whitespace only method", func(t *testing.T) {
		body := `{"model":"sys.session","method":"   ","args":[]}`
		resp, status := Dispatch(ctx, []byte(body))
		if status != http.StatusBadRequest || resp.Error.Code != CodeValidationError {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})

	t.Run("unsupported public method default", func(t *testing.T) {
		PublicMethods["__test_only__"] = true
		defer delete(PublicMethods, "__test_only__")
		resp, status := Dispatch(ctx, []byte(`{"model":"sys.session","method":"__test_only__","args":[]}`))
		if status != http.StatusForbidden || resp.Error.Code != CodeMethodNotAllowed {
			t.Fatalf("status=%d resp=%+v", status, resp)
		}
	})
}

func TestValidateKwargs(t *testing.T) {
	if err := validateKwargs(nil); err != nil {
		t.Fatal(err)
	}
	if err := validateKwargs([]byte("null")); err != nil {
		t.Fatal(err)
	}
	err := validateKwargs([]byte(`[]`))
	if err == nil {
		t.Fatal("expected error for array kwargs")
	}
}

package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"sumeru/core/server/api"
)

func TestSuccess_and_Fail(t *testing.T) {
	ok := api.Success([]int{1})
	if !ok.OK || ok.Error != nil || ok.Result == nil {
		t.Fatalf("Success: %+v", ok)
	}
	fail := api.Fail(api.CodeValidationError, "bad", map[string]interface{}{"field": "x"})
	if fail.OK || fail.Error == nil || fail.Error.Code != api.CodeValidationError {
		t.Fatalf("Fail: %+v", fail)
	}
}

func TestWriteResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	api.WriteResponse(rr, http.StatusBadRequest, api.Fail(api.CodeInvalidArgs, "nope", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["ok"] != false {
		t.Fatalf("ok = %v", m["ok"])
	}
}

func TestResponseFromError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		resp, status := api.ResponseFromError(nil)
		if status != http.StatusOK || !resp.OK {
			t.Fatalf("resp=%+v status=%d", resp, status)
		}
	})
	t.Run("coded", func(t *testing.T) {
		resp, status := api.ResponseFromError(api.NewRPCErrorForTest(api.CodeUnauthorized, "auth", nil))
		if status != http.StatusUnauthorized || resp.Error.Code != api.CodeUnauthorized {
			t.Fatalf("resp=%+v status=%d", resp, status)
		}
	})
	t.Run("classified", func(t *testing.T) {
		resp, status := api.ResponseFromError(errors.New("unknown model foo"))
		if status != http.StatusNotFound || resp.Error.Code != api.CodeModelNotFound {
			t.Fatalf("resp=%+v status=%d", resp, status)
		}
	})
}

func TestStatusForCode(t *testing.T) {
	cases := map[string]int{
		api.CodeUnauthorized:         http.StatusUnauthorized,
		api.CodeInvalidArgs:          http.StatusBadRequest,
		api.CodeValidationError:      http.StatusBadRequest,
		api.CodeUnsupportedMediaType: http.StatusUnsupportedMediaType,
		api.CodePayloadTooLarge:      http.StatusRequestEntityTooLarge,
		api.CodeMethodNotAllowed:     http.StatusForbidden,
		api.CodeAccessDenied:         http.StatusForbidden,
		api.CodeModelNotFound:        http.StatusNotFound,
		api.CodeNotFound:             http.StatusNotFound,
		api.CodeInternalError:        http.StatusInternalServerError,
		"UNKNOWN":                    http.StatusInternalServerError,
	}
	for code, want := range cases {
		if got := api.StatusForCodeForTest(code); got != want {
			t.Fatalf("StatusForCodeForTest(%q) = %d; want %d", code, got, want)
		}
	}
}

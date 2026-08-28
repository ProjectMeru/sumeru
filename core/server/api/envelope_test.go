package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccess_and_Fail(t *testing.T) {
	ok := Success([]int{1})
	if !ok.OK || ok.Error != nil || ok.Result == nil {
		t.Fatalf("Success: %+v", ok)
	}
	fail := Fail(CodeValidationError, "bad", map[string]interface{}{"field": "x"})
	if fail.OK || fail.Error == nil || fail.Error.Code != CodeValidationError {
		t.Fatalf("Fail: %+v", fail)
	}
}

func TestWriteResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteResponse(rr, http.StatusBadRequest, Fail(CodeInvalidArgs, "nope", nil))
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
		resp, status := ResponseFromError(nil)
		if status != http.StatusOK || !resp.OK {
			t.Fatalf("resp=%+v status=%d", resp, status)
		}
	})
	t.Run("coded", func(t *testing.T) {
		resp, status := ResponseFromError(newRPCError(CodeUnauthorized, "auth", nil))
		if status != http.StatusUnauthorized || resp.Error.Code != CodeUnauthorized {
			t.Fatalf("resp=%+v status=%d", resp, status)
		}
	})
	t.Run("classified", func(t *testing.T) {
		resp, status := ResponseFromError(errors.New("unknown model foo"))
		if status != http.StatusNotFound || resp.Error.Code != CodeModelNotFound {
			t.Fatalf("resp=%+v status=%d", resp, status)
		}
	})
}

func TestStatusForCode(t *testing.T) {
	cases := map[string]int{
		CodeUnauthorized:         http.StatusUnauthorized,
		CodeInvalidArgs:          http.StatusBadRequest,
		CodeValidationError:      http.StatusBadRequest,
		CodeUnsupportedMediaType: http.StatusUnsupportedMediaType,
		CodePayloadTooLarge:      http.StatusRequestEntityTooLarge,
		CodeMethodNotAllowed:     http.StatusForbidden,
		CodeAccessDenied:         http.StatusForbidden,
		CodeModelNotFound:        http.StatusNotFound,
		CodeNotFound:             http.StatusNotFound,
		CodeInternalError:        http.StatusInternalServerError,
		"UNKNOWN":                http.StatusInternalServerError,
	}
	for code, want := range cases {
		if got := statusForCode(code); got != want {
			t.Fatalf("statusForCode(%q) = %d; want %d", code, got, want)
		}
	}
}

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"sumeru/core/server/api"
)

func TestDispatch_inputValidation(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		body       string
		wantStatus int
		wantOK     bool
		wantCode   string
	}{
		{
			name:       "invalid json",
			ctx:        authCtx(),
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeInvalidJSON,
		},
		{
			name:       "empty body",
			ctx:        authCtx(),
			body:       ``,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeInvalidJSON,
		},
		{
			name:       "missing model",
			ctx:        authCtx(),
			body:       `{"method":"search"}`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeValidationError,
		},
		{
			name:       "missing method",
			ctx:        authCtx(),
			body:       `{"model":"core.user"}`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeValidationError,
		},
		{
			name:       "unknown model",
			ctx:        authCtx(),
			body:       `{"model":"no.such.model","method":"search","args":[]}`,
			wantStatus: http.StatusNotFound,
			wantOK:     false,
			wantCode:   api.CodeModelNotFound,
		},
		{
			name:       "disallowed method",
			ctx:        authCtx(),
			body:       `{"model":"sys.session","method":"execute","args":[]}`,
			wantStatus: http.StatusForbidden,
			wantOK:     false,
			wantCode:   api.CodeMethodNotAllowed,
		},
		{
			name:       "args not array",
			ctx:        authCtx(),
			body:       `{"model":"sys.session","method":"search","args":{}}`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "kwargs not object",
			ctx:        authCtx(),
			body:       `{"model":"sys.session","method":"search","args":[],"kwargs":[]}`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "invalid params wrapper",
			ctx:        authCtx(),
			body:       `{"params":{`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeInvalidJSON,
		},
		{
			name:       "params wrapper missing model",
			ctx:        authCtx(),
			body:       `{"params":{"method":"search","args":[]}}`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeValidationError,
		},
		{
			name:       "unauthenticated",
			ctx:        context.Background(),
			body:       `{"model":"sys.session","method":"search","args":[]}`,
			wantStatus: http.StatusUnauthorized,
			wantOK:     false,
			wantCode:   api.CodeUnauthorized,
		},
		{
			name:       "search_read missing args",
			ctx:        authCtx(),
			body:       `{"model":"sys.session","method":"search_read","args":[[]]}`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "read missing ids",
			ctx:        authCtx(),
			body:       `{"model":"sys.session","method":"read","args":[]}`,
			wantStatus: http.StatusBadRequest,
			wantOK:     false,
			wantCode:   api.CodeInvalidArgs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, status := api.Dispatch(tt.ctx, []byte(tt.body))
			if status != tt.wantStatus {
				t.Fatalf("status = %d; want %d", status, tt.wantStatus)
			}
			if resp.OK != tt.wantOK {
				t.Fatalf("ok = %v; want %v", resp.OK, tt.wantOK)
			}
			if tt.wantOK {
				if resp.Error != nil {
					t.Fatalf("error = %+v; want nil", resp.Error)
				}
				return
			}
			if resp.Error == nil {
				t.Fatal("error is nil")
			}
			if resp.Error.Code != tt.wantCode {
				t.Fatalf("error.code = %q; want %q (msg=%q)", resp.Error.Code, tt.wantCode, resp.Error.Message)
			}
			if resp.Result != nil {
				t.Fatalf("result = %v; want nil on failure", resp.Result)
			}
		})
	}
}

func TestDispatch_paramsWrapper(t *testing.T) {
	body := `{"params":{"model":"no.such.model","method":"search","args":[]}}`
	resp, status := api.Dispatch(authCtx(), []byte(body))
	if status != http.StatusNotFound {
		t.Fatalf("status = %d; want %d", status, http.StatusNotFound)
	}
	if resp.OK || resp.Error == nil || resp.Error.Code != api.CodeModelNotFound {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestDispatch_successEnvelope(t *testing.T) {
	// Dispatch returns ok:true with null result only when dispatch succeeds; without DB
	// we only assert envelope shape on validation paths. Verify Success helper shape via JSON.
	raw, err := json.Marshal(api.Success([]int{1, 2, 3}))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["ok"] != true {
		t.Fatalf("ok = %v", m["ok"])
	}
	if m["error"] != nil {
		t.Fatalf("error = %v; want null", m["error"])
	}
}

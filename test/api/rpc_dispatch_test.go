package api_test

import (
	"net/http"
	"strings"
	"testing"

	"sumeru/core/server/api"
)

func TestDispatch_methodValidation(t *testing.T) {
	model := "sys.session"
	tests := []struct {
		name       string
		method     string
		args       string
		kwargs     string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "create missing values",
			method:     "create",
			args:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "create invalid values object",
			method:     "create",
			args:       `[[]]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "write missing args",
			method:     "write",
			args:       `[[]]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "write invalid ids",
			method:     "write",
			args:       `[{},{}]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "unlink missing ids",
			method:     "unlink",
			args:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "unlink invalid ids",
			method:     "unlink",
			args:       `[{}]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "create_many missing list",
			method:     "create_many",
			args:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "create_many invalid list",
			method:     "create_many",
			args:       `[{}]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "write_many missing values",
			method:     "write_many",
			args:       `[[1]]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "write_many invalid ids",
			method:     "write_many",
			args:       `[{},{}]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "unlink_many missing ids",
			method:     "unlink_many",
			args:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "onchange missing args",
			method:     "onchange",
			args:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "onchange invalid values",
			method:     "onchange",
			args:       `[[],1]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "onchange empty field",
			method:     "onchange",
			args:       `[{},""]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "read_group missing spec",
			method:     "read_group",
			args:       `[]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "call missing method",
			method:     "call",
			args:       `[1]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "call invalid id",
			method:     "call",
			args:       `["x","foo"]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "call empty method name",
			method:     "call",
			args:       `[1,""]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "read invalid ids type",
			method:     "read",
			args:       `[{}]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "read invalid fields",
			method:     "read",
			args:       `[[1],{}]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "search_read invalid fields",
			method:     "search_read",
			args:       `[[],{}]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
		{
			name:       "search invalid domain",
			method:     "search",
			args:       `[["bad"]]`,
			wantStatus: http.StatusBadRequest,
			wantCode:   api.CodeInvalidArgs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kwargs := tt.kwargs
			if kwargs == "" {
				kwargs = `{}`
			}
			body := `{"model":"` + model + `","method":"` + tt.method + `","args":` + tt.args + `,"kwargs":` + kwargs + `}`
			assertDispatch(t, authCtx(), body, tt.wantStatus, tt.wantCode)
		})
	}
}

func TestDispatch_readEmptyIDs(t *testing.T) {
	body := `{"model":"sys.session","method":"read","args":[[]]}`
	resp, status := api.Dispatch(authCtx(), []byte(body))
	if status != http.StatusOK || !resp.OK {
		t.Fatalf("status=%d resp=%+v", status, resp)
	}
}

func TestDispatch_readTooManyIDs(t *testing.T) {
	ids := make([]string, 0, 501)
	for i := 1; i <= 501; i++ {
		ids = append(ids, "1")
	}
	body := `{"model":"sys.session","method":"read","args":[` + strings.Join(ids, ",") + `]}`
	assertDispatch(t, authCtx(), body, http.StatusBadRequest, api.CodeInvalidArgs)
}

func TestDispatch_writeUnlinkEmptyIDs(t *testing.T) {
	for _, body := range []string{
		`{"model":"sys.session","method":"write","args":[[],{"sid":"x"}]}`,
		`{"model":"sys.session","method":"unlink","args":[[]]}`,
	} {
		resp, status := api.Dispatch(authCtx(), []byte(body))
		if status != http.StatusOK || !resp.OK || resp.Result != true {
			t.Fatalf("body=%s status=%d resp=%+v", body, status, resp)
		}
	}
}

func TestDispatch_writeManyEmptyIDs(t *testing.T) {
	body := `{"model":"sys.session","method":"write_many","args":[[],{"name":"x"}]}`
	resp, status := api.Dispatch(authCtx(), []byte(body))
	if status != http.StatusOK {
		t.Fatalf("status = %d; want 200 for empty id list", status)
	}
	if !resp.OK || resp.Result != true {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestDispatch_unlinkManyEmptyIDs(t *testing.T) {
	body := `{"model":"sys.session","method":"unlink_many","args":[[]]}`
	resp, status := api.Dispatch(authCtx(), []byte(body))
	if status != http.StatusOK {
		t.Fatalf("status = %d; want 200 for empty id list", status)
	}
	if !resp.OK || resp.Result != true {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestDispatch_readGroupInvalidDomain(t *testing.T) {
	body := `{"model":"sys.session","method":"read_group","args":[{"domain":"not-a-domain","groupby":["id"]}]}`
	assertDispatch(t, authCtx(), body, http.StatusBadRequest, api.CodeInvalidArgs)
}

func TestDispatch_createManyTooManyRows(t *testing.T) {
	rows := make([]string, 0, 501)
	for i := 0; i < 501; i++ {
		rows = append(rows, `{}`)
	}
	body := `{"model":"sys.session","method":"create_many","args":[` + strings.Join(rows, ",") + `]}`
	assertDispatch(t, authCtx(), body, http.StatusBadRequest, api.CodeInvalidArgs)
}
